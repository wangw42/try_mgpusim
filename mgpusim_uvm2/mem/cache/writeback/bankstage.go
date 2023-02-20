package writeback

import (
	"log"

	"gitlab.com/akita/akita/v2/sim"
	"gitlab.com/akita/mem/v2/mem"
	"gitlab.com/akita/util/v2/pipelining"
	"gitlab.com/akita/util/v2/tracing"
)

type bankStage struct {
	cache  *Cache
	bankID int

	pipeline                   pipelining.Pipeline
	pipelineWidth              int
	postPipelineBuf            *bufferImpl
	inflightTransCount         int
	downwardInflightTransCount int // Count the trans that needs to be sent to the write buffer.
}

type bufferImpl struct {
	capacity int
	elements []interface{}
}

func (b *bufferImpl) CanPush() bool {
	return len(b.elements) < b.capacity
}

func (b *bufferImpl) Push(e interface{}) {
	if len(b.elements) >= b.capacity {
		log.Panic("buffer overflow")
	}
	b.elements = append(b.elements, e)
}

func (b *bufferImpl) Pop() interface{} {
	if len(b.elements) == 0 {
		return nil
	}

	e := b.elements[0]
	b.elements = b.elements[1:]
	return e
}

func (b *bufferImpl) Peek() interface{} {
	if len(b.elements) == 0 {
		return nil
	}

	return b.elements[0]
}

func (b *bufferImpl) Capacity() int {
	return b.capacity
}

func (b *bufferImpl) Size() int {
	return len(b.elements)
}

func (b *bufferImpl) Clear() {
	b.elements = nil
}

func (b *bufferImpl) Get(i int) interface{} {
	return b.elements[i]
}

func (b *bufferImpl) Remove(i int) {
	b.elements = append(b.elements[:i], b.elements[i+1:]...)
}

type bankPipelineElem struct {
	trans *transaction
}

func (e bankPipelineElem) TaskID() string {
	return e.trans.req().Meta().ID + "_write_back_bank_pipeline"
}

func (s *bankStage) Tick(now sim.VTimeInSec) (madeProgress bool) {
	for i := 0; i < s.cache.numReqPerCycle; i++ {
		madeProgress = s.finalizeTrans(now) || madeProgress
	}

	madeProgress = s.pipeline.Tick(now) || madeProgress

	for i := 0; i < s.cache.numReqPerCycle; i++ {
		madeProgress = s.pullFromBuf(now) || madeProgress
	}

	return madeProgress
}

func (s *bankStage) Reset(now sim.VTimeInSec) {
	s.cache.dirToBankBuffers[s.bankID].Clear()
	// s.currentTrans = nil
}

func (s *bankStage) pullFromBuf(now sim.VTimeInSec) bool {
	if !s.pipeline.CanAccept() {
		return false
	}

	inBuf := s.cache.writeBufferToBankBuffers[s.bankID]
	trans := inBuf.Pop()
	if trans != nil {
		s.pipeline.Accept(now, bankPipelineElem{trans: trans.(*transaction)})
		s.inflightTransCount++
		return true
	}

	// Do not jam the writeBufferBuffer
	if !s.cache.writeBufferBuffer.CanPush() {
		return false
	}

	// Always reserve one lane for up-going transactions
	if s.downwardInflightTransCount >= s.pipelineWidth-1 {
		return false
	}

	inBuf = s.cache.dirToBankBuffers[s.bankID]
	trans = inBuf.Pop()
	if trans != nil {
		t := trans.(*transaction)

		if t.action == writeBufferFetch {
			s.cache.writeBufferBuffer.Push(trans)
			return true
		}

		s.pipeline.Accept(now, bankPipelineElem{trans: trans.(*transaction)})
		s.inflightTransCount++

		switch t.action {
		case bankEvict, bankEvictAndFetch, bankEvictAndWrite:
			s.downwardInflightTransCount++
		}

		return true
	}

	return false
}

func (s *bankStage) finalizeTrans(now sim.VTimeInSec) bool {
	for i := 0; i < s.postPipelineBuf.Size(); i++ {
		trans := s.postPipelineBuf.Get(i).(bankPipelineElem).trans

		done := false

		switch trans.action {
		case bankReadHit:
			done = s.finalizeReadHit(now, trans)
		case bankWriteHit:
			done = s.finalizeWriteHit(now, trans)
		case bankWriteFetched:
			done = s.finalizeBankWriteFetched(now, trans)
		case bankEvictAndFetch, bankEvictAndWrite, bankEvict:
			done = s.finalizeBankEviction(now, trans)
		default:
			panic("bank action not supported")
		}

		if done {
			s.postPipelineBuf.Remove(i)

			return true
		}
	}

	return false
}

func (s *bankStage) finalizeReadHit(
	now sim.VTimeInSec,
	trans *transaction,
) bool {
	if !s.cache.topSender.CanSend(1) {
		return false
	}

	read := trans.read
	addr := read.Address
	_, offset := getCacheLineID(addr, s.cache.log2BlockSize)
	block := trans.block

	data, err := s.cache.storage.Read(
		block.CacheAddress+offset, read.AccessByteSize)
	if err != nil {
		panic(err)
	}

	s.removeTransaction(trans)
	s.inflightTransCount--
	s.downwardInflightTransCount--
	block.ReadCount--

	dataReady := mem.DataReadyRspBuilder{}.
		WithSendTime(now).
		WithSrc(s.cache.topPort).
		WithDst(read.Src).
		WithRspTo(read.ID).
		WithData(data).
		Build()
	s.cache.topSender.Send(dataReady)

	tracing.TraceReqComplete(read, now, s.cache)

	return true
}

func (s *bankStage) finalizeWriteHit(
	now sim.VTimeInSec,
	trans *transaction,
) bool {
	if !s.cache.topSender.CanSend(1) {
		return false
	}

	write := trans.write
	addr := write.Address
	_, offset := getCacheLineID(addr, s.cache.log2BlockSize)
	block := trans.block

	data, err := s.cache.storage.Read(
		block.CacheAddress, 1<<s.cache.log2BlockSize)
	if err != nil {
		panic(err)
	}
	dirtyMask := block.DirtyMask
	if dirtyMask == nil {
		dirtyMask = make([]bool, 1<<s.cache.log2BlockSize)
	}
	for i := 0; i < len(write.Data); i++ {
		if write.DirtyMask == nil || write.DirtyMask[i] {
			index := offset + uint64(i)
			data[index] = write.Data[i]
			dirtyMask[index] = true
		}
	}
	err = s.cache.storage.Write(block.CacheAddress, data)
	if err != nil {
		panic(err)
	}

	block.IsValid = true
	block.IsLocked = false
	block.IsDirty = true
	block.DirtyMask = dirtyMask

	s.removeTransaction(trans)
	s.inflightTransCount--
	s.downwardInflightTransCount--

	done := mem.WriteDoneRspBuilder{}.
		WithSendTime(now).
		WithSrc(s.cache.topPort).
		WithDst(write.Src).
		WithRspTo(write.ID).
		Build()
	s.cache.topSender.Send(done)

	tracing.TraceReqComplete(write, now, s.cache)

	return true
}

func (s *bankStage) finalizeBankWriteFetched(
	now sim.VTimeInSec,
	trans *transaction,
) bool {
	if !s.cache.mshrStageBuffer.CanPush() {
		return false
	}

	mshrEntry := trans.mshrEntry
	block := mshrEntry.Block
	s.cache.mshrStageBuffer.Push(mshrEntry)
	s.cache.storage.Write(block.CacheAddress, mshrEntry.Data)
	block.IsLocked = false
	block.IsValid = true

	s.inflightTransCount--

	return true
}

func (s *bankStage) removeTransaction(trans *transaction) {
	for i, t := range s.cache.inFlightTransactions {
		if trans == t {
			s.cache.inFlightTransactions = append(
				(s.cache.inFlightTransactions)[:i],
				(s.cache.inFlightTransactions)[i+1:]...)
			return
		}
	}
	panic("transaction not found")
}

func (s *bankStage) finalizeBankEviction(
	now sim.VTimeInSec,
	trans *transaction,
) bool {
	if !s.cache.writeBufferBuffer.CanPush() {
		return false
	}

	victim := trans.victim
	data, err := s.cache.storage.Read(
		victim.CacheAddress, 1<<s.cache.log2BlockSize)
	if err != nil {
		panic(err)
	}
	trans.evictingData = data

	switch trans.action {
	case bankEvict:
		trans.action = writeBufferFlush
	case bankEvictAndFetch:
		trans.action = writeBufferEvictAndFetch
	case bankEvictAndWrite:
		trans.action = writeBufferEvictAndWrite
	default:
		panic("unsupported action")
	}

	delete(s.cache.evictingList, trans.evictingAddr)
	s.cache.writeBufferBuffer.Push(trans)
	s.inflightTransCount--
	s.downwardInflightTransCount--

	return true
}

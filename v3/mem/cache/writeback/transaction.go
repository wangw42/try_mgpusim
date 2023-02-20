package writeback

import (
	"gitlab.com/akita/akita/v3/sim"
	"gitlab.com/akita/mem/v3/cache"
	"gitlab.com/akita/mem/v3/mem"
	"gitlab.com/akita/mem/v3/vm"
)

type action int

const (
	actionInvalid action = iota
	bankReadHit
	bankWriteHit
	bankEvict
	bankEvictAndWrite
	bankEvictAndFetch
	bankWriteFetched
	writeBufferFetch
	writeBufferEvictAndFetch
	writeBufferEvictAndWrite
	writeBufferFlush
)

type transaction struct {
	action
	id                string
	read              *mem.ReadReq
	write             *mem.WriteReq
	flush             *cache.FlushReq
	block             *cache.Block
	victim            *cache.Block
	fetchPID          vm.PID
	fetchAddress      uint64
	fetchedData       []byte
	fetchReadReq      *mem.ReadReq
	evictingPID       vm.PID
	evictingAddr      uint64
	evictingData      []byte
	evictingDirtyMask []bool
	evictionWriteReq  *mem.WriteReq
	evictionDone      *mem.WriteDoneRsp
	mshrEntry         *cache.MSHREntry
}

func (t transaction) accessReq() mem.AccessReq {
	if t.read != nil {
		return t.read
	}
	if t.write != nil {
		return t.write
	}
	return nil
}

func (t transaction) req() sim.Msg {
	if t.accessReq() != nil {
		return t.accessReq()
	}
	if t.flush != nil {
		return t.flush
	}
	return nil
}

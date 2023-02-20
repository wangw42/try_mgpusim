package writeback

import (
	"gitlab.com/akita/akita/v2/sim"
	"gitlab.com/akita/mem/v2/mem"
	"gitlab.com/akita/util/v2/tracing"
)

type topParser struct {
	cache *Cache
}

func (p *topParser) Tick(now sim.VTimeInSec) bool {
	if p.cache.state != cacheStateRunning {
		return false
	}

	req := p.cache.topPort.Peek()
	if req == nil {
		return false
	}

	if !p.cache.dirStageBuffer.CanPush() {
		return false
	}

	trans := &transaction{
		id: sim.GetIDGenerator().Generate(),
	}
	switch req := req.(type) {
	case *mem.ReadReq:
		trans.read = req
	case *mem.WriteReq:
		trans.write = req
	}
	p.cache.dirStageBuffer.Push(trans)

	p.cache.inFlightTransactions = append(p.cache.inFlightTransactions, trans)

	tracing.TraceReqReceive(req, now, p.cache)

	p.cache.topPort.Retrieve(now)

	return true
}

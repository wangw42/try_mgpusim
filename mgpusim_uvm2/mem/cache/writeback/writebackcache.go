package writeback

import (
	"gitlab.com/akita/akita/v2/sim"
	"gitlab.com/akita/mem/v2/cache"
	"gitlab.com/akita/mem/v2/mem"
	"gitlab.com/akita/util/v2/akitaext"
	"gitlab.com/akita/util/v2/buffering"
)

type cacheState int

const (
	cacheStateInvalid cacheState = iota
	cacheStateRunning
	cacheStatePreFlushing
	cacheStateFlushing
	cacheStatePaused
)

// A Cache in the writeback package is a cache that performs the write-back policy.
type Cache struct {
	*sim.TickingComponent

	topPort     sim.Port
	bottomPort  sim.Port
	controlPort sim.Port

	dirStageBuffer           buffering.Buffer
	dirToBankBuffers         []buffering.Buffer
	writeBufferToBankBuffers []buffering.Buffer
	mshrStageBuffer          buffering.Buffer
	writeBufferBuffer        buffering.Buffer

	topSender         akitaext.BufferedSender
	bottomSender      akitaext.BufferedSender
	controlPortSender akitaext.BufferedSender

	topParser   *topParser
	writeBuffer *writeBufferStage
	dirStage    *directoryStage
	bankStages  []*bankStage
	mshrStage   *mshrStage
	flusher     *flusher

	storage         *mem.Storage
	lowModuleFinder mem.LowModuleFinder
	directory       cache.Directory
	mshr            cache.MSHR
	log2BlockSize   uint64
	numReqPerCycle  int

	state                cacheState
	inFlightTransactions []*transaction
	evictingList         map[uint64]bool
}

// SetLowModuleFinder sets the LowModuleFinder used by the cache.
func (c *Cache) SetLowModuleFinder(lmf mem.LowModuleFinder) {
	c.lowModuleFinder = lmf
}

// Tick updates the internal states of the Cache.
func (c *Cache) Tick(now sim.VTimeInSec) bool {
	madeProgress := false

	madeProgress = c.controlPortSender.Tick(now) || madeProgress

	if c.state != cacheStatePaused {
		madeProgress = c.runPipeline(now) || madeProgress
	}

	madeProgress = c.flusher.Tick(now) || madeProgress

	return madeProgress
}

func (c *Cache) runPipeline(now sim.VTimeInSec) bool {
	madeProgress := false

	madeProgress = c.runStage(now, c.topSender) || madeProgress
	madeProgress = c.runStage(now, c.bottomSender) || madeProgress
	madeProgress = c.runStage(now, c.mshrStage) || madeProgress

	for _, bs := range c.bankStages {
		madeProgress = bs.Tick(now) || madeProgress
	}

	madeProgress = c.runStage(now, c.writeBuffer) || madeProgress
	madeProgress = c.runStage(now, c.dirStage) || madeProgress
	madeProgress = c.runStage(now, c.topParser) || madeProgress

	return madeProgress
}

func (c *Cache) runStage(now sim.VTimeInSec, stage sim.Ticker) bool {
	madeProgress := false
	for i := 0; i < c.numReqPerCycle; i++ {
		madeProgress = stage.Tick(now) || madeProgress
	}
	return madeProgress
}

func (c *Cache) discardInflightTransactions(now sim.VTimeInSec) {
	sets := c.directory.GetSets()
	for _, set := range sets {
		for _, block := range set.Blocks {
			block.ReadCount = 0
			block.IsLocked = false
		}
	}

	c.dirStage.Reset(now)
	for _, bs := range c.bankStages {
		bs.Reset(now)
	}
	c.mshrStage.Reset(now)
	c.writeBuffer.Reset(now)

	clearPort(c.topPort, now)

	c.topSender.Clear()

	c.inFlightTransactions = nil
}

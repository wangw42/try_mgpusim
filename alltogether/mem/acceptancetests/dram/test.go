package main

import (
	"flag"
	"fmt"
	"math/rand"

	"gitlab.com/akita/akita/v2/sim"
	"gitlab.com/akita/mem/v2/dram"

	"os"
	"time"

	"log"

	"gitlab.com/akita/mem/v2/acceptancetests"
	"gitlab.com/akita/mem/v2/trace"
	"gitlab.com/akita/util/v2/tracing"
)

var seedFlag = flag.Int64("seed", 0, "Random Seed")
var numAccessFlag = flag.Int("num-access", 100000, "Number of accesses to generate")
var maxAddressFlag = flag.Uint64("max-address", 1048576, "Address range to use")
var traceFileFlag = flag.String("trace", "", "Trace file")
var parallelFlag = flag.Bool("parallel", false, "Test with parallel engine")

func main() {
	flag.Parse()

	var seed int64
	if *seedFlag == 0 {
		seed = time.Now().UnixNano()
	} else {
		seed = *seedFlag
	}
	fmt.Fprintf(os.Stderr, "Seed %d\n", seed)
	rand.Seed(seed)

	var engine sim.Engine
	if *parallelFlag {
		engine = sim.NewParallelEngine()
	} else {
		engine = sim.NewSerialEngine()
	}
	//engine.AcceptHook(sim.NewEventLogger(log.New(os.Stdout, "", 0)))

	conn := sim.NewDirectConnection("conn", engine, 1*sim.GHz)

	agent := acceptancetests.NewMemAccessAgent(engine)
	agent.MaxAddress = *maxAddressFlag
	agent.WriteLeft = *numAccessFlag
	agent.ReadLeft = *numAccessFlag

	memCtrl := dram.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		Build("mem")

	agent.LowModule = memCtrl.GetPortByName("Top")

	if *traceFileFlag != "" {
		traceFile, _ := os.Create(*traceFileFlag)
		logger := log.New(traceFile, "", 0)
		tracer := trace.NewTracer(logger)
		tracing.CollectTrace(memCtrl, tracer)
	}

	conn.PlugIn(agent.GetPortByName("Mem"), 16)
	conn.PlugIn(memCtrl.GetPortByName("Top"), 1)

	agent.TickLater(0)

	err := engine.Run()
	if err != nil {
		panic(err)
	}

	if len(agent.PendingWriteReq) > 0 || len(agent.PendingReadReq) > 0 {
		panic("Not all req returned")
	}

	if agent.WriteLeft > 0 || agent.ReadLeft > 0 {
		panic("more requests to send")
	}
}

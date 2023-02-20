package org

import (
	"gitlab.com/akita/akita/v2/sim"
	"gitlab.com/akita/mem/v2/dram/internal/signal"
	"gitlab.com/akita/util/v2/tracing"
)

// A Bank is a DRAM Bank. It contains a number of rows and columns.
type Bank interface {
	tracing.NamedHookable

	GetReadyCommand(
		now sim.VTimeInSec,
		cmd *signal.Command,
	) *signal.Command
	StartCommand(now sim.VTimeInSec, cmd *signal.Command)
	UpdateTiming(cmdKind signal.CommandKind, cycleNeeded int)
	Tick(now sim.VTimeInSec) bool
}

// Package cmdq provides command queue implementations
package cmdq

import (
	"gitlab.com/akita/akita/v2/sim"
	"gitlab.com/akita/mem/v2/dram/internal/signal"
)

// A CommandQueue is a queue of command that needs to be executed by a rank or
// a bank.
type CommandQueue interface {
	GetCommandToIssue(
		now sim.VTimeInSec,
	) *signal.Command
	CanAccept(command *signal.Command) bool
	Accept(command *signal.Command)
}

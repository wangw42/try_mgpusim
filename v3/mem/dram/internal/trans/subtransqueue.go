package trans

import (
	"gitlab.com/akita/akita/v3/sim"
	"gitlab.com/akita/mem/v3/dram/internal/signal"
)

// A SubTransactionQueue is a queue for subtransactions.
type SubTransactionQueue interface {
	CanPush(n int) bool
	Push(t *signal.Transaction)
	Tick(now sim.VTimeInSec) bool
}

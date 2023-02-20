package trans

import (
	"gitlab.com/akita/akita/v2/sim"
	"gitlab.com/akita/mem/v2/dram/internal/signal"
)

// A SubTransactionQueue is a queue for subtransactions.
type SubTransactionQueue interface {
	CanPush(n int) bool
	Push(t *signal.Transaction)
	Tick(now sim.VTimeInSec) bool
}

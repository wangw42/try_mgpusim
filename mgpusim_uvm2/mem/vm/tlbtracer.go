package vm

import (
	"fmt"
	"io"

	"gitlab.com/akita/akita/v2/sim"
)

// A TLBTracer write logs for what happened in a TLB
type TLBTracer struct {
	writer io.Writer
}

// NewTLBTracer produce a new TLBTracer, injecting the dependency of a writer.
func NewTLBTracer(w io.Writer) *TLBTracer {
	t := new(TLBTracer)
	t.writer = w
	return t
}

// Func prints the tlb trace information.
func (t *TLBTracer) Func(ctx *sim.HookCtx) {
	what, ok := ctx.Item.(string)
	if !ok {
		return
	}

	_, err := fmt.Fprintf(t.writer,
		"%.12f,%s,%s,{}\n",
		ctx.Now,
		ctx.Domain.(sim.Component).Name(),
		what)
	if err != nil {
		panic(err)
	}
}

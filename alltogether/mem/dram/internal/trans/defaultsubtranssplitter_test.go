package trans

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.com/akita/mem/v2/dram/internal/signal"
	"gitlab.com/akita/mem/v2/mem"
)

var _ = Describe("Default SubTransSplitter", func() {

	It("should split", func() {
		read := mem.ReadReqBuilder{}.
			WithAddress(1020).
			WithByteSize(128).
			Build()
		transaction := &signal.Transaction{
			Read: read,
		}

		splitter := NewSubTransSplitter(6)

		splitter.Split(transaction)

		Expect(transaction.SubTransactions).To(HaveLen(3))
	})
})

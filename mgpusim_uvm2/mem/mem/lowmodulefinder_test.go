package mem

import (
	"fmt"

	"gitlab.com/akita/akita/v2/sim"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InterleavedLowModuleFinder", func() {
	var (
		lowModuleFinder *InterleavedLowModuleFinder
	)

	BeforeEach(func() {
		lowModuleFinder = new(InterleavedLowModuleFinder)
		lowModuleFinder.UseAddressSpaceLimitation = true
		lowModuleFinder.LowAddress = 0
		lowModuleFinder.HighAddress = 4 * GB
		lowModuleFinder.InterleavingSize = 4096
		lowModuleFinder.LowModules = make([]sim.Port, 0)
		for i := 0; i < 6; i++ {
			lowModuleFinder.LowModules = append(
				lowModuleFinder.LowModules,
				sim.NewLimitNumMsgPort(nil, 4,
					fmt.Sprintf("LowModule_%d.Port", i)))
		}
		lowModuleFinder.ModuleForOtherAddresses =
			sim.NewLimitNumMsgPort(nil, 4, "LowModule_other.Port")
	})

	It("should find low module if address is in-space", func() {
		Expect(lowModuleFinder.Find(0)).To(
			BeIdenticalTo(lowModuleFinder.LowModules[0]))
		Expect(lowModuleFinder.Find(4096)).To(
			BeIdenticalTo(lowModuleFinder.LowModules[1]))
		Expect(lowModuleFinder.Find(4097)).To(
			BeIdenticalTo(lowModuleFinder.LowModules[1]))
	})

	It("should use a special module for all the addresses that does not fall in range", func() {
		Expect(lowModuleFinder.Find(4 * GB)).To(
			BeIdenticalTo(lowModuleFinder.ModuleForOtherAddresses))
	})
})

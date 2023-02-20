package tlb

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -destination "mock_sim_test.go" -package $GOPACKAGE -write_package_comment=false gitlab.com/akita/akita/v3/sim Port,Engine,BufferedSender
//go:generate mockgen -destination "mock_mem_test.go" -package $GOPACKAGE -write_package_comment=false gitlab.com/akita/mem/v3/mem LowModuleFinder
//go:generate mockgen -destination "mock_internal_test.go" -package $GOPACKAGE -write_package_comment=false gitlab.com/akita/mem/v3/vm/tlb/internal Set
func TestTlb(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tlb Suite")
}

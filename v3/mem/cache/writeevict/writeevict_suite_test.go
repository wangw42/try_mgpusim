package writeevict

import (
	"log"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -destination "mock_cache_test.go" -package $GOPACKAGE -write_package_comment=false gitlab.com/akita/mem/v3/cache Directory,MSHR
//go:generate mockgen -destination "mock_mem_test.go" -package $GOPACKAGE -write_package_comment=false gitlab.com/akita/mem/v3/mem LowModuleFinder
//go:generate mockgen -destination "mock_sim_test.go" -package $GOPACKAGE -write_package_comment=false gitlab.com/akita/akita/v3/sim Port,Buffer
//go:generate mockgen -destination "mock_pipelinging_test.go" -package $GOPACKAGE -write_package_comment=false "gitlab.com/akita/akita/v3/pipelining"  Pipeline
func TestWriteevict(t *testing.T) {
	log.SetOutput(GinkgoWriter)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Writeevict Suite")
}

package addresstranslator

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -destination "mock_sim_test.go" -package $GOPACKAGE -write_package_comment=false gitlab.com/akita/akita/v2/sim Port,Engine
//go:generate mockgen -destination "mock_mem_test.go" -package $GOPACKAGE -write_package_comment=false gitlab.com/akita/mem/v2/mem LowModuleFinder
//go:generate mockgen -destination "mock_akitaext_test.go" -package $GOPACKAGE -write_package_comment=false gitlab.com/akita/util/v2/akitaext BufferedSender
func TestAddresstranslator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Address Translator Suite")
}

package idealmemcontroller

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -destination "mock_sim_test.go" -package $GOPACKAGE -write_package_comment=false gitlab.com/akita/akita/v3/sim Port,Connection,Engine
func TestIdealmemcontroller(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Idealmemcontroller Suite")
}

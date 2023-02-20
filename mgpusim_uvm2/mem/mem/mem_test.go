package mem

import (
	"log"
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

func TestMemory(t *testing.T) {
	log.SetOutput(ginkgo.GinkgoWriter)
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Mem")
}

package dram

import (
	"testing"

	"gitlab.com/akita/akita/v2/sim"
	"gitlab.com/akita/mem/v2/mem"

	"github.com/golang/mock/gomock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -destination "mock_sim_test.go" -package $GOPACKAGE -write_package_comment=false gitlab.com/akita/akita/v2/sim Port
//go:generate mockgen -destination "mock_trans_test.go" -package $GOPACKAGE -write_package_comment=false gitlab.com/akita/mem/v2/dram/internal/trans SubTransactionQueue,SubTransSplitter
//go:generate mockgen -destination "mock_addressmapping_test.go" -package $GOPACKAGE -write_package_comment=false gitlab.com/akita/mem/v2/dram/internal/addressmapping AddressConverter,Mapper
//go:generate mockgen -destination "mock_cmdq_test.go" -package $GOPACKAGE -write_package_comment=false gitlab.com/akita/mem/v2/dram/internal/cmdq CommandQueue
//go:generate mockgen -destination "mock_org_test.go" -package $GOPACKAGE -write_package_comment=false gitlab.com/akita/mem/v2/dram/internal/org Channel

func TestDram(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dram Suite")
}

var _ = Describe("DRAM Integration", func() {
	var (
		mockCtrl *gomock.Controller
		engine   sim.Engine
		srcPort  *MockPort
		memCtrl  *MemController
		conn     *sim.DirectConnection
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		engine = sim.NewSerialEngine()
		memCtrl = MakeBuilder().
			WithEngine(engine).
			Build("memCtrl")
		srcPort = NewMockPort(mockCtrl)

		conn = sim.NewDirectConnection("conn", engine, 1*sim.GHz)
		srcPort.EXPECT().SetConnection(conn)
		conn.PlugIn(memCtrl.topPort, 1)
		conn.PlugIn(srcPort, 1)
	})

	It("should read and write", func() {
		write := mem.WriteReqBuilder{}.
			WithAddress(0x40).
			WithData([]byte{1, 2, 3, 4}).
			WithSrc(srcPort).
			WithDst(memCtrl.topPort).
			WithSendTime(0).
			Build()

		read := mem.ReadReqBuilder{}.
			WithAddress(0x40).
			WithByteSize(4).
			WithSrc(srcPort).
			WithDst(memCtrl.topPort).
			WithSendTime(0).
			Build()

		memCtrl.topPort.Recv(write)
		memCtrl.topPort.Recv(read)

		ret1 := srcPort.EXPECT().
			Recv(gomock.Any()).
			Do(func(wd *mem.WriteDoneRsp) {
				Expect(wd.RespondTo).To(Equal(write.ID))
			})
		srcPort.EXPECT().
			Recv(gomock.Any()).
			Do(func(dr *mem.DataReadyRsp) {
				Expect(dr.RespondTo).To(Equal(read.ID))
				Expect(dr.Data).To(Equal([]byte{1, 2, 3, 4}))
			}).After(ret1)

		engine.Run()
	})
})

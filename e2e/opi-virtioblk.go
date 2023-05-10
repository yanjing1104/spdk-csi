package e2e

import (
	"time"

	"github.com/ShellCode33/VM-Detection/vmdetect"
	ginkgo "github.com/onsi/ginkgo/v2"
	"k8s.io/kubernetes/test/e2e/framework"
)

const (
	opiVirtioBlkConfigMapData = `{
  "nodes": [
    {
      "name": "localhost",
      "rpcURL": "http://127.0.0.1:9009",
      "targetType": "nvme-tcp",
      "targetAddr": "127.0.0.1"
    }
  ]
}`
)

var _ = ginkgo.Describe("SPDKCSI-OPI-VirtioBlk", func() {
	f := framework.NewDefaultFramework("spdkcsi")
	ginkgo.BeforeEach(func() {
		deployConfigs(opiVirtioBlkConfigMapData)
		deployOpiVirtioBlkConfig()
		deployCsi()
	})

	ginkgo.AfterEach(func() {
		deleteCsi()
		deleteOpiVirtioBlkConfig()
		deleteConfigs()
	})

	ginkgo.Context("Test SPDK CSI OPI VirtioBlk", func() {
		ginkgo.It("Test SPDK CSI OPI VirtioBlk", func() {
			if isVM, _ := vmdetect.IsRunningInVirtualMachine(); !isVM {
				ginkgo.Skip("Skipping SPDKCSI-OPI-VirtioBlk test: Running inside a virtual machine")
			}

			ginkgo.By("checking controller statefulset is running", func() {
				err := waitForControllerReady(f.ClientSet, 4*time.Minute)
				if err != nil {
					ginkgo.Fail(err.Error())
				}
			})

			ginkgo.By("checking node daemonset is running", func() {
				err := waitForNodeServerReady(f.ClientSet, 2*time.Minute)
				if err != nil {
					ginkgo.Fail(err.Error())
				}
			})

			ginkgo.By("log verification for xPU grpc connection", func() {
				expLogerrMsgMap := map[string]string{
					"connected to xPU node 127.0.0.1:50051 with TargetType as xpu-opi-virtioblk": "failed to catch the log about the connection to xPU node",
				}
				err := verifyNodeServerLog(expLogerrMsgMap)
				if err != nil {
					ginkgo.Fail(err.Error())
				}
			})

			ginkgo.By("create multiple pvcs and a pod with multiple pvcs attached, and check data persistence after the pod is removed and recreated", func() {
				deployMultiPvcs()
				deployTestPodWithMultiPvcs()
				err := waitForTestPodReady(f.ClientSet, 5*time.Minute)
				if err != nil {
					ginkgo.Fail(err.Error())
				}

				err = checkDataPersistForMultiPvcs(f)
				if err != nil {
					ginkgo.Fail(err.Error())
				}

				deleteMultiPvcsAndTestPodWithMultiPvcs()
				err = waitForTestPodGone(f.ClientSet)
				if err != nil {
					ginkgo.Fail(err.Error())
				}
				for _, pvcName := range []string{"spdkcsi-pvc1", "spdkcsi-pvc2", "spdkcsi-pvc3"} {
					err = waitForPvcGone(f.ClientSet, pvcName)
					if err != nil {
						ginkgo.Fail(err.Error())
					}
				}
			})

			ginkgo.By("log verification for OPI workflow", func() {
				expLogerrMsgMap := map[string]string{
					"OPI.CreateNVMfRemoteController": "failed to catch the log about the OPI.CreateNVMfRemoteController phase",
					"OPI.CreateVirtioBlk":            "failed to catch the log about the OPI.CreateVirtioBlk phase",
					"OPI.DeleteVirtioBlk":            "failed to catch the log about the OPI.DeleteVirtioBlk phase",
					"OPI.DeleteNVMfRemoteController": "failed to catch the log about the OPI.DeleteNVMfRemoteController phase",
				}
				err := verifyNodeServerLog(expLogerrMsgMap)
				if err != nil {
					ginkgo.Fail(err.Error())
				}
			})
		})
	})
})

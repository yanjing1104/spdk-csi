/*
Copyright (c) Intel Corporation.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"k8s.io/klog"

	opiapiCommon "github.com/opiproject/opi-api/common/v1/gen/go"
	opiapiStorage "github.com/opiproject/opi-api/storage/v1alpha1/gen/go"
)

const (
	opiNvmfRemoteControllerHostnqnPref = "nqn.2023-04.io.spdk.csi:remote.controller:uuid:"
	opiNVMeSubsystemNqnPref            = "nqn.2016-06.io.spdk.csi:subsystem:uuid:"
)

func NewSpdkCsiOpiInitiator(volumeContext map[string]string, xpuConnClient grpc.ClientConnInterface, xpuTargetType string, kvmPciBridges int) (SpdkCsiInitiator, error) {
	iOpiCommon := &opiCommon{
		opiClient:     xpuConnClient,
		volumeContext: volumeContext,
		timeout:       60 * time.Second,
		kvmPciBridges: kvmPciBridges,
	}
	switch xpuTargetType {
	case "xpu-opi-virtioblk":
		return &opiInitiatorVirtioBlk{opi: iOpiCommon}, nil
	case "xpu-opi-nvme":
		return &opiInitiatorNvme{opi: iOpiCommon}, nil
	default:
		return nil, fmt.Errorf("unknown xPU targetType: %s", xpuTargetType)
	}
}

type opiCommon struct {
	volumeContext              map[string]string
	timeout                    time.Duration
	kvmPciBridges              int
	opiClient                  grpc.ClientConnInterface
	nvmfRemoteControllerClient opiapiStorage.NVMfRemoteControllerServiceClient
	nvmfRemoteControllerID     string
	devicePath                 string
}

type opiInitiatorNvme struct {
	opi                *opiCommon
	frontendNvmeClient opiapiStorage.FrontendNvmeServiceClient
	nvmeControllerID   string
	nvmeSubsystemID    string
	nvmeNamespaceID    string
}

type opiInitiatorVirtioBlk struct {
	opi                     *opiCommon
	frontendVirtioBlkClient opiapiStorage.FrontendVirtioBlkServiceClient
	virtioBlkID             string
}

func (opi *opiCommon) ctxTimeout() (context.Context, context.CancelFunc) {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), opi.timeout)
	return ctxTimeout, cancel
}

// Connect to remote controller, which is needed for by OPI VirtioBlk and Nvme
func (opi *opiCommon) createNvmfRemoteController(nvmfRemoteControllerID string) error {
	ctxTimeout, cancel := opi.ctxTimeout()
	defer cancel()

	// Check if NVMfRemoteController exists
	// If it does not exists, there will be err for GetNVMfRemoteController: rpc error: code = Unknown desc = bdev_nvme_get_controllers: \
	// json response error: Controller remote.controller.spdk.csi:08b58258-2be9-4d31-9522-ae8707d56cde does not exist
	getReq := &opiapiStorage.GetNVMfRemoteControllerRequest{
		Name: nvmfRemoteControllerID,
	}
	klog.Infof("OPI.GetNVMfRemoteControllerRequest(...) => %s", getReq)

	getRes, err := opi.nvmfRemoteControllerClient.GetNVMfRemoteController(ctxTimeout, getReq)
	if err != nil {
		klog.Infof("OPI.GetNVMfRemoteController error: %s", err)
	}
	klog.Info("OPI.GetNVMfRemoteController(...) => ", getRes)

	// Create NVMfRemoteController if it does not exists
	if getRes == nil && strings.Contains(err.Error(), "does not exist") {
		trsvcid, err := strconv.ParseInt(opi.volumeContext["targetPort"], 10, 64)
		if err != nil {
			klog.Errorf("get trsvcid error: %s", err)
		}
		createReq := &opiapiStorage.CreateNVMfRemoteControllerRequest{
			NvMfRemoteController: &opiapiStorage.NVMfRemoteController{
				Id: &opiapiCommon.ObjectKey{
					Value: nvmfRemoteControllerID,
				},
				Trtype:  4,
				Adrfam:  1,
				Traddr:  opi.volumeContext["targetAddr"],
				Trsvcid: trsvcid,
				Subnqn:  opi.volumeContext["nqn"],
				Hostnqn: opiNvmfRemoteControllerHostnqnPref + opi.volumeContext["model"],
			},
		}
		klog.Infof("OPI.CreateNVMfRemoteControllerRequest(...) => %s", createReq)
		createRes, err := opi.nvmfRemoteControllerClient.CreateNVMfRemoteController(ctxTimeout, createReq)
		if err != nil {
			return fmt.Errorf("OPI.CreateNVMfRemoteController error: %w", err)
		}
		opi.nvmfRemoteControllerID = createRes.Id.Value
		klog.Info("OPI.CreateNVMfRemoteController(...) => ", createRes)
	}

	return nil
}

// Disconnect from remote controller, which is needed by both OPI VirtioBlk and Nvme
func (opi *opiCommon) deleteNvmfRemoteController(nvmfRemoteControllerID string) error {
	ctxTimeout, cancel := opi.ctxTimeout()
	defer cancel()

	// DeleteNVMfRemoteController, with "AllowMissing: true", deleting operation will always succeed even the resource is not found
	deleteReq := &opiapiStorage.DeleteNVMfRemoteControllerRequest{
		Name:         nvmfRemoteControllerID,
		AllowMissing: true,
	}
	klog.Infof("OPI.DeleteNVMfRemoteControllerRequest(...) => %s", deleteReq)

	_, err := opi.nvmfRemoteControllerClient.DeleteNVMfRemoteController(ctxTimeout, deleteReq)
	if err != nil {
		return fmt.Errorf("OPI.DeleteNVMfRemoteController error: %w", err)
	}
	klog.Info("OPI.DeleteNVMfRemoteController successfully")

	return nil
}

// Create a controller with VirtioBlk transport information Bdev
func (i *opiInitiatorVirtioBlk) createVirtioBlk() (bdf string, err error) {
	ctxTimeout, cancel := i.opi.ctxTimeout()
	defer cancel()

	createVirtioBlkFlag := false

	for pciBridge := 1; pciBridge <= i.opi.kvmPciBridges; pciBridge++ {
		for busCount := 0; busCount < 32; busCount++ {
			physID := int32(busCount + 32*(pciBridge-1))
			virtioBlkID := "virtioblk." + strconv.Itoa(int(physID))

			// Check if the controller with VirtioBlk transport information Bdev
			// If it does not exists, there will be err for GetVirtioBlk: rpc error: code = InvalidArgument desc = Could not find Controller: virtioblk.0
			getReq := &opiapiStorage.GetVirtioBlkRequest{
				Name: virtioBlkID,
			}
			klog.Infof("OPI.GetVirtioBlkRequest(...) => %s", getReq)

			getRes, err := i.frontendVirtioBlkClient.GetVirtioBlk(ctxTimeout, getReq)
			if err != nil {
				klog.Errorf("OPI.GetVirtioBlk error: %s", err)
			}
			klog.Info("OPI.GetVirtioBlk(...) => ", getRes)

			// Create a controller with VirtioBlk transport information Bdev
			if getRes == nil && strings.Contains(err.Error(), "Could not find Controller") {
				createReq := &opiapiStorage.CreateVirtioBlkRequest{
					VirtioBlk: &opiapiStorage.VirtioBlk{
						Id: &opiapiCommon.ObjectKey{
							Value: virtioBlkID,
						},
						PcieId: &opiapiStorage.PciEndpoint{
							PhysicalFunction: physID,
						},
						VolumeId: &opiapiCommon.ObjectKey{
							Value: i.opi.volumeContext["model"],
						},
					},
				}
				klog.Infof("OPI.CreateVirtioBlkRequest(...) => %s", createReq)
				createRes, err := i.frontendVirtioBlkClient.CreateVirtioBlk(ctxTimeout, createReq)
				if err != nil {
					klog.Errorf("OPI.CreateVirtioBlk with pfId (%d) error: %s", physID, err)
				} else {
					createVirtioBlkFlag = true
					i.virtioBlkID = createRes.Id.Value
					bdf = fmt.Sprintf("0000:%02d:%02x.0", pciBridge, busCount)
					klog.Info("OPI.CreateVirtioBlk(...) => ", createRes)
					break
				}
			}
		}
		if createVirtioBlkFlag {
			break
		}
	}

	if !createVirtioBlkFlag {
		klog.Errorf("OPI.CreateVirtioBlk failed with all PhysicalIds")
		return "", fmt.Errorf("could not CreateVirtioBlk for OPI.VirtioBlk")
	}

	return bdf, nil
}

// Delete the controller with VirtioBlk transport information Bdev
func (i *opiInitiatorVirtioBlk) deleteVirtioBlk(virtioBlkID string) error {
	ctxTimeout, cancel := i.opi.ctxTimeout()
	defer cancel()

	// DeleteVirtioBlk, with "AllowMissing: true", deleting operation will always succeed even the resource is not found
	deleteReq := &opiapiStorage.DeleteVirtioBlkRequest{
		Name:         virtioBlkID,
		AllowMissing: true,
	}
	klog.Infof("OPI.DeleteVirtioBlkRequest(...) => %s", deleteReq)

	_, err := i.frontendVirtioBlkClient.DeleteVirtioBlk(ctxTimeout, deleteReq)
	if err != nil {
		klog.Errorf("OPI.Nvme DeleteVirtioBlk error: %s", err)
		return err
	}
	klog.Info("OPI.DeleteVirtioBlk successfully")

	return nil
}

// Create the subsystem
func (i *opiInitiatorNvme) createNVMeSubsystem(nvmeSubsystemID string) error {
	ctxTimeout, cancel := i.opi.ctxTimeout()
	defer cancel()

	// Check if the subsystem exists
	// If it does not exist, there will be err for GetNVMeSubsystem: rpc error: code = NotFound desc = \
	// unable to find key nvme.subsystem.spdk.csi:9a8d71be-4e37-4c32-8da9-b71fbcf198a1
	getReq := &opiapiStorage.GetNVMeSubsystemRequest{
		Name: nvmeSubsystemID,
	}
	klog.Infof("OPI.GetNVMeSubsystemRequest(...) => %s", getReq)

	getRes, err := i.frontendNvmeClient.GetNVMeSubsystem(ctxTimeout, getReq)
	if err != nil {
		klog.Infof("OPI.GetNVMeSubsystem error: %s", err)
	}
	klog.Info("OPI.GetNVMeSubsystem(...) => ", getRes)

	// Create the subsystem if it does not exist
	if getRes == nil && strings.Contains(err.Error(), "unable to find key") {
		createReq := &opiapiStorage.CreateNVMeSubsystemRequest{
			NvMeSubsystem: &opiapiStorage.NVMeSubsystem{
				Spec: &opiapiStorage.NVMeSubsystemSpec{
					Id: &opiapiCommon.ObjectKey{
						Value: nvmeSubsystemID,
					},
					Nqn: opiNVMeSubsystemNqnPref + i.opi.volumeContext["model"],
				},
			},
		}
		klog.Infof("OPI.CreateNVMeSubsystemRequest(...) => %s", createReq)

		createRes, err := i.frontendNvmeClient.CreateNVMeSubsystem(ctxTimeout, createReq)
		if err != nil {
			return fmt.Errorf("OPI.CreateNVMeSubsystem error: %w", err)
		}
		klog.Info("OPI.CreateNVMeSubsystem(...) => ", createRes)

		i.nvmeSubsystemID = createRes.Spec.Id.Value
	}

	return nil
}

// deleteNVMeSubsystem
func (i *opiInitiatorNvme) deleteNVMeSubsystem(nvmeSubsystemID string) error {
	ctxTimeout, cancel := i.opi.ctxTimeout()
	defer cancel()

	// Delete the subsystem, with "AllowMissing: true", deleting operation will always succeed even the resource is not found
	deleteReq := &opiapiStorage.DeleteNVMeSubsystemRequest{
		Name:         nvmeSubsystemID,
		AllowMissing: true,
	}
	klog.Infof("OPI.DeleteNVMeSubsystemRequest(...) => %s", deleteReq)

	_, err := i.frontendNvmeClient.DeleteNVMeSubsystem(ctxTimeout, deleteReq)
	if err != nil {
		return fmt.Errorf("OPI.DeleteNVMeSubsystem error: %w", err)
	}
	klog.Info("OPI.DeleteNVMeSubsystem successfully")

	return nil
}

// Create a controller with vfiouser transport information for Nvme
func (i *opiInitiatorNvme) createNVMeController() (bdf string, err error) {
	ctxTimeout, cancel := i.opi.ctxTimeout()
	defer cancel()

	createNVMeControllerFlag := false

	// Traverse all the possible physIDs until createNVMeController succeeds
	for pciBridge := 1; pciBridge <= i.opi.kvmPciBridges; pciBridge++ {
		for busCount := 0; busCount < 32; busCount++ {
			physID := int32(busCount + 32*(pciBridge-1))
			nvmeControllerID := "nvmecontroller." + strconv.Itoa(int(physID))

			// Check if the controller with vfiouser transport information exists
			// If it does not exist, there will be err for GetNVMeController: rpc error: code = NotFound desc = unable to find key nvmecontroller.0
			getReq := &opiapiStorage.GetNVMeControllerRequest{
				Name: nvmeControllerID,
			}
			klog.Infof("OPI.GetNVMeControllerRequest(...) => %s", getReq)

			getRes, err := i.frontendNvmeClient.GetNVMeController(ctxTimeout, getReq)
			if err != nil {
				klog.Errorf("OPI.GetNVMeController error: %s", err)
			}
			klog.Info("OPI.GetNVMeController(...) => ", getRes)

			// Create the controller with vfiouser transport information if it does not exist
			if getRes == nil && strings.Contains(err.Error(), "unable to find key") {
				createReq := &opiapiStorage.CreateNVMeControllerRequest{
					NvMeController: &opiapiStorage.NVMeController{
						Spec: &opiapiStorage.NVMeControllerSpec{
							Id: &opiapiCommon.ObjectKey{
								Value: nvmeControllerID,
							},
							NvmeControllerId: 0,
							SubsystemId: &opiapiCommon.ObjectKey{
								Value: i.nvmeSubsystemID,
							},

							PcieId: &opiapiStorage.PciEndpoint{
								PhysicalFunction: physID,
							},
						},
					},
				}
				klog.Infof("OPI.CreateNVMeControllerRequest(...) => %s", createReq)

				createRes, err := i.frontendNvmeClient.CreateNVMeController(ctxTimeout, createReq)
				if err != nil {
					klog.Errorf("OPI.CreateNVMeController with pfId (%d) error: %s", physID, err)
				} else {
					createNVMeControllerFlag = true
					bdf = fmt.Sprintf("0000:%02d:%02x.0", pciBridge, busCount)
					i.nvmeControllerID = createRes.Spec.Id.Value
					klog.Info("OPI.CreateNVMeController(...) with pfId => ", physID, createRes)
					break
				}
			}
		}
		if createNVMeControllerFlag {
			break
		}
	}

	if !createNVMeControllerFlag {
		klog.Errorf("OPI.CreateNVMeController failed with all PhysicalIds")
		return "", fmt.Errorf("could not CreateNVMeController for OPI.Nvme")
	}

	return bdf, nil
}

// Delete the controller with vfiouser transport information for Nvme
func (i *opiInitiatorNvme) deleteNVMeController(nvmeControllerName string) (err error) {
	ctxTimeout, cancel := i.opi.ctxTimeout()
	defer cancel()

	// Delete the controller with vfiouser transport information, with "AllowMissing: true", deleting operation will always succeed even the resource is not found
	deleteControllerReq := &opiapiStorage.DeleteNVMeControllerRequest{
		Name:         nvmeControllerName,
		AllowMissing: true,
	}
	klog.Infof("OPI.DeleteNVMeControllerRequest(...) => %s", deleteControllerReq)

	_, err = i.frontendNvmeClient.DeleteNVMeController(ctxTimeout, deleteControllerReq)
	if err != nil {
		klog.Errorf("OPI.Nvme DeleteNVMeController error: %s", err)
		return err
	}
	klog.Info("OPI.DeleteNVMeController successfully")

	return nil
}

// Get Bdev for the volume and add a new namespace to the subsystem with that bdev for Nvme
func (i *opiInitiatorNvme) createNVMeNamespace(nvmeNamespaceID string) error {
	ctxTimeout, cancel := i.opi.ctxTimeout()
	defer cancel()

	// Check if the namespace associated with the volumeID exists
	// if it does not exist, there will be err for GetNVMeNamespace: rpc error: code = NotFound desc = unable to find key nvme.namespace.spdk.csi:16099265-d405-4946-a709-3666b9ca13b7
	getReq := &opiapiStorage.GetNVMeNamespaceRequest{
		Name: nvmeNamespaceID,
	}
	klog.Infof("OPI.GetNVMeNamespaceRequest(...) => %s", getReq)

	getRes, err := i.frontendNvmeClient.GetNVMeNamespace(ctxTimeout, getReq)
	if err != nil {
		klog.Errorf("OPI.GetNVMeNamespace error: %s", err)
	}
	klog.Info("OPI.GetNVMeNamespace(...) => ", getRes)

	// get Bdev for the volume and add a new namespace to the subsystem with that bdev if the namespace associated with the volumeID does not exist

	if getRes == nil && strings.Contains(err.Error(), "unable to find key") {
		createReq := &opiapiStorage.CreateNVMeNamespaceRequest{
			NvMeNamespace: &opiapiStorage.NVMeNamespace{
				Spec: &opiapiStorage.NVMeNamespaceSpec{
					Id: &opiapiCommon.ObjectKey{
						Value: nvmeNamespaceID,
					},
					SubsystemId: &opiapiCommon.ObjectKey{
						Value: i.nvmeSubsystemID,
					},
					VolumeId: &opiapiCommon.ObjectKey{
						Value: i.opi.volumeContext["model"],
					},
				},
			},
		}
		klog.Infof("OPI.CreateNVMeNamespaceRequest(...) => %s", createReq)

		createRes, err := i.frontendNvmeClient.CreateNVMeNamespace(ctxTimeout, createReq)
		if err != nil {
			return err
		}
		klog.Info("OPI.CreateNVMeNamespace(...) => ", createRes)

		i.nvmeNamespaceID = createRes.Spec.Id.Value
	}

	return nil
}

// Delete the namespace from the subsystem with the bdev for Nvme
func (i *opiInitiatorNvme) deleteNVMeNamespace(nvmeNamespaceID string) error {
	ctxTimeout, cancel := i.opi.ctxTimeout()
	defer cancel()

	// Delete namespace, with "AllowMissing: true", deleting operation will always succeed even the resource is not found
	deleteNvmeNamespaceReq := &opiapiStorage.DeleteNVMeNamespaceRequest{
		Name:         nvmeNamespaceID,
		AllowMissing: true,
	}
	klog.Infof("OPI.DeleteNVMeNamespaceRequest(...) => %s", deleteNvmeNamespaceReq)

	_, err := i.frontendNvmeClient.DeleteNVMeNamespace(ctxTimeout, deleteNvmeNamespaceReq)
	if err != nil {
		klog.Errorf("OPI.Nvme DeleteNVMeNamespace error: %s", err)
		return err
	}
	klog.Info("OPI.DeleteNVMeNamespace successfully")

	return nil
}

// get the Nvme block device
func (i *opiInitiatorNvme) getNvmeDevice(bdf string) (string, error) {
	// find the uuid file path for the nvme device based on the bdf
	uuidFilePath, err := waitForDeviceReady(fmt.Sprintf("/sys/bus/pci/devices/%s/nvme/nvme*/nvme*n*/uuid", bdf))
	if err != nil {
		return "", fmt.Errorf("OPI.Nvme wait for the uuid file (%s) error: %w", uuidFilePath, err)
	}
	klog.Infof("uuidFilePath is %s", uuidFilePath)

	// verify whether the uuid file's content match the volumeID
	uuidContent, err := os.ReadFile(uuidFilePath)
	if err != nil {
		return "", fmt.Errorf("OPI.Nvme open the uuid file (%s) error: %w", uuidFilePath, err)
	}
	klog.Infof("uuidContent is %s, volumeID is %s", strings.TrimSpace(string(uuidContent)), i.opi.volumeContext["model"])

	deviceName := ""
	if strings.TrimSpace(string(uuidContent)) == i.opi.volumeContext["model"] {
		// Obtain the part nvme*c*n* or nvme*n* from the file path
		deviceSysFileName := NvmeReDeviceSysFileName.FindString(uuidFilePath)
		// Remove c* from (nvme*c*n*)
		deviceName = NvmeReDeviceName.ReplaceAllString(deviceSysFileName, "")
		klog.Infof("deviceSysFileName is %s, deviceName is %s", deviceSysFileName, deviceName)
	}

	if deviceName == "" {
		return "", fmt.Errorf("OPI.Nvme uuid file's content does not match volume ID")
	}

	deviceGlob := fmt.Sprintf("/dev/%s", deviceName)
	devicePath, err := waitForDeviceReady(deviceGlob)
	if err != nil {
		return "", err
	}
	klog.Infof("Device path is %s", devicePath)

	return devicePath, nil
}

// cleanup for OPI Nvme
func (i *opiInitiatorNvme) cleanup() {
	// All the deleting operations have "AllowMissing: true" in the request, they will always succeed even the resources are not found
	// So te cleanup contains all the resources deleting operations
	if i.nvmeSubsystemID != "" {
		if err := i.deleteNVMeSubsystem(i.nvmeSubsystemID); err != nil {
			klog.Info("OPI.Nvme workflow failed, call Delete* to clean up err: ", err)
		}
	}

	if i.nvmeControllerID != "" {
		if err := i.deleteNVMeController(i.nvmeControllerID); err != nil {
			klog.Info("OPI.Nvme workflow failed, call Delete* to clean up err:", err)
		}
	}

	if i.opi.nvmfRemoteControllerID != "" {
		if err := i.opi.deleteNvmfRemoteController(i.opi.nvmfRemoteControllerID); err != nil {
			klog.Info("OPI.Nvme workflow failed, call Delete* to clean up err:", err)
		}
	}

	if i.nvmeNamespaceID != "" {
		if err := i.deleteNVMeNamespace(i.nvmeNamespaceID); err != nil {
			klog.Info("OPI.Nvme workflow failed, call Delete* to clean up err:", err)
		}
	}
}

// For OPI Nvme Connect(), four steps will be included.
// The first step is Create a new subsystem, the nqn (nqn.2016-06.io.spdk.csi:subsystem:uuid:VolumeId) will be set in the CreateNVMeSubsystemRequest.
// After a successful CreateNVMeSubsystem, a nvmf subsystem with the nqn will be created in the xPU node.
// The second step is create a controller with vfiouser transport information, we are using KVM case now,
// and the only information needed in the CreateNVMeControllerRequest is pfId
// which should be from 0 to the sum of buses-counts (namely 64 in our case). After a successful CreateNVMeController, the "listen_addresses" field in the nvmf subsystem
// created in the first step will be filled in with VFIOUSER related information,
// including transport (VFIOUSER), trtype (VFIOUSER), adrfam (IPv4) and traddr (/var/tmp/controller$pfId).
// The third step is to connect to the remote controller, this step is used to connect to the storage node.
// The last step is to get Bdev for the volume and add a new namespace to the subsystem with that bdev. After this step, the Nvme block device will appear.
// If any step above fails, call cleanup operation to clean the resources.
func (i *opiInitiatorNvme) Connect() (string, error) {
	i.opi.nvmfRemoteControllerClient = opiapiStorage.NewNVMfRemoteControllerServiceClient(i.opi.opiClient)
	i.frontendNvmeClient = opiapiStorage.NewFrontendNvmeServiceClient(i.opi.opiClient)

	// step 1: create a subsystem
	err := i.createNVMeSubsystem("nvme.subsystem.spdk.csi:" + i.opi.volumeContext["model"])
	if err != nil {
		return "", err
	}

	// step 2: create a controller with vfiouser transport information
	bdf, err := i.createNVMeController()
	if err != nil {
		i.cleanup()
		return "", err
	}

	// step 3: connect to remote controller
	err = i.opi.createNvmfRemoteController("remote.controller.spdk.csi:" + i.opi.volumeContext["model"])
	if err != nil {
		i.cleanup()
		return "", err
	}

	// step 4: get Bdev for the volume and add a new namespace to the subsystem with that bdev
	err = i.createNVMeNamespace("nvme.namespace.spdk.csi:" + i.opi.volumeContext["model"])
	if err != nil {
		i.cleanup()
		return "", err
	}

	// wait for the block device ready
	devicePath, err := i.getNvmeDevice(bdf)
	if err != nil {
		i.cleanup()
		return "", err
	}

	i.opi.devicePath = devicePath

	return devicePath, nil
}

// For OPI Nvme Disconnect(), three steps will be included, namely DeleteNVMfRemoteController, DeleteNVMeController and DeleteNVMeSubsystem.
// DeleteNVMeNamespace is skipped cause when deleting system, namespace will be deleted automatically
func (i *opiInitiatorNvme) Disconnect() error {
	// step 1: deleteNVMfRemoteController
	if err := i.opi.deleteNvmfRemoteController(i.opi.nvmfRemoteControllerID); err != nil {
		return err
	}

	// step 2: deleteNVMeController if it exists
	if err := i.deleteNVMeController(i.nvmeControllerID); err != nil {
		return err
	}

	// step 3: deleteNVMeSubsystem
	if err := i.deleteNVMeSubsystem(i.nvmeSubsystemID); err != nil {
		return err
	}

	return waitForDeviceGone(i.opi.devicePath)
}

// cleanup for OPI VirtioBlk
func (i *opiInitiatorVirtioBlk) cleanup() {
	// All the deleting operations have "AllowMissing: true" in the request, they will always succeed even the resources are not found
	// So te cleanup contains all the resources deleting operations

	if i.opi.nvmfRemoteControllerID != "" {
		if err := i.opi.deleteNvmfRemoteController(i.opi.nvmfRemoteControllerID); err != nil {
			klog.Info("OPI.VirtioBlk workflow failed, call Delete* to clean up err:", err)
		}
	}

	if i.virtioBlkID != "" {
		if err := i.deleteVirtioBlk(i.virtioBlkID); err != nil {
			klog.Info("OPI.VirtioBlk workflow failed, call Delete* to clean up err:", err)
		}
	}
}

// For OPI VirtioBlk Connect(), two steps will be included.
// The first step is to connect to the remote controller, this step is used to connect to the storage node.
// The second step is CreateVirtioBlk, which is calling vhost_create_blk_controller on xPU node.
// After these two steps, a VirtioBlk device will appear.
func (i *opiInitiatorVirtioBlk) Connect() (string, error) {
	i.opi.nvmfRemoteControllerClient = opiapiStorage.NewNVMfRemoteControllerServiceClient(i.opi.opiClient)
	i.frontendVirtioBlkClient = opiapiStorage.NewFrontendVirtioBlkServiceClient(i.opi.opiClient)

	// step 1: connect to remote controller
	err := i.opi.createNvmfRemoteController("remote.controller.spdk.csi:" + i.opi.volumeContext["model"])
	if err != nil {
		return "", err
	}

	// step 2: Create a controller with virtio_blk transport information Bdev
	bdf, err := i.createVirtioBlk()
	if err != nil {
		i.cleanup()
		return "", err
	}

	// find the dir file path for the virtio device
	devicePath, err := GetVirtioBlkDevice(bdf)
	if err != nil {
		i.cleanup()
		return "", err
	}

	i.opi.devicePath = devicePath

	return devicePath, nil
}

// For OPI VirtioBlk Disconnect(), two steps will be included, namely DeleteVirtioBlk and DeleteNVMfRemoteController.
func (i *opiInitiatorVirtioBlk) Disconnect() error {
	// DeleteVirtioBlk if it exists
	err := i.deleteVirtioBlk(i.virtioBlkID)
	if err != nil {
		return err
	}

	// DeleteNVMfRemoteController
	err = i.opi.deleteNvmfRemoteController(i.opi.nvmfRemoteControllerID)
	if err != nil {
		return err
	}

	return waitForDeviceGone(i.opi.devicePath)
}

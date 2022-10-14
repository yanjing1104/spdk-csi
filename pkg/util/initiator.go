/*
Copyright (c) Arm Limited and Contributors.

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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"github.com/google/uuid"
	"k8s.io/klog"

	// grpc stuff

	smarpc "spdk.io/sma"
	"spdk.io/sma/nvmf_tcp"

	spdkcsiConfig "github.com/spdk/spdk-csi/pkg/config"
)

// SpdkCsiInitiator defines interface for NVMeoF/iSCSI/SMA initiator
//   - Connect initiates target connection and returns local block device filename
//     e.g., /dev/disk/by-id/nvme-SPDK_Controller1_SPDK00000000000001
//   - Disconnect terminates target connection
//   - Caller(node service) should serialize calls to same initiator
//   - Implementation should be idempotent to duplicated requests
type SpdkCsiInitiator interface {
	Connect() (string, error)
	Disconnect() error
}

func NewSpdkCsiInitiator(volumeContext map[string]string) (SpdkCsiInitiator, error) {
	targetType := strings.ToLower(volumeContext["targetType"])
	if smaConfigString, ok := volumeContext["sma"]; ok {
		klog.Infof("SMA in volumeContext, use legacy connection to %v", volumeContext["model"])
		isma := initiatorSMA{volumeContext: volumeContext}
		smaConfig := spdkcsiConfig.SmaConfig{}
		err := json.Unmarshal([]byte(smaConfigString), &smaConfig)
		if err != nil {
			return nil, fmt.Errorf("invalid SMA configuration: %q (%w)", smaConfigString, err)
		}
		isma.serverURL = smaConfig.Server
		switch targetType {
		case "tcp":
			isma.req = &smarpc.CreateDeviceRequest{
				Volume: nil,
				Params: &smarpc.CreateDeviceRequest_NvmfTcp{
					NvmfTcp: &nvmf_tcp.DeviceParameters{
						Subnqn:  volumeContext["nqn"],
						Adrfam:  "ipv4",
						Traddr:  volumeContext["targetAddr"],
						Trsvcid: "4420",
					},
				},
			}
		default:
			klog.Errorf("Unsupported SMA target type in %v", volumeContext)
			return nil, fmt.Errorf("unknown SMA target type: %q", volumeContext["targetType"])
		}
		return &isma, nil
	}
	klog.Infof("No SMA in volumeContext, use legacy connection to %v", volumeContext)
	switch targetType {
	case "rdma", "tcp":
		return &initiatorNVMf{
			// see util/nvmf.go VolumeInfo()
			targetType: volumeContext["targetType"],
			targetAddr: volumeContext["targetAddr"],
			targetPort: volumeContext["targetPort"],
			nqn:        volumeContext["nqn"],
			model:      volumeContext["model"],
		}, nil
	case "iscsi":
		return &initiatorISCSI{
			targetAddr: volumeContext["targetAddr"],
			targetPort: volumeContext["targetPort"],
			iqn:        volumeContext["iqn"],
		}, nil
	default:
		return nil, fmt.Errorf("unknown initiator: %s", targetType)
	}
}

// NVMf initiator implementation
type initiatorNVMf struct {
	targetType string
	targetAddr string
	targetPort string
	nqn        string
	model      string
}

func (nvmf *initiatorNVMf) Connect() (string, error) {
	// nvme connect -t tcp -a 192.168.1.100 -s 4420 -n "nqn"
	cmdLine := []string{
		"nvme", "connect", "-t", strings.ToLower(nvmf.targetType),
		"-a", nvmf.targetAddr, "-s", nvmf.targetPort, "-n", nvmf.nqn,
	}
	err := execWithTimeout(cmdLine, 40)
	if err != nil {
		// go on checking device status in case caused by duplicated request
		klog.Errorf("command %v failed: %s", cmdLine, err)
	}

	deviceGlob := fmt.Sprintf("/dev/disk/by-id/*%s*", nvmf.model)
	devicePath, err := waitForDeviceReady(deviceGlob, 20)
	if err != nil {
		return "", err
	}
	return devicePath, nil
}

func (nvmf *initiatorNVMf) Disconnect() error {
	// nvme disconnect -n "nqn"
	cmdLine := []string{"nvme", "disconnect", "-n", nvmf.nqn}
	err := execWithTimeout(cmdLine, 40)
	if err != nil {
		// go on checking device status in case caused by duplicate request
		klog.Errorf("command %v failed: %s", cmdLine, err)
	}

	deviceGlob := fmt.Sprintf("/dev/disk/by-id/*%s*", nvmf.model)
	return waitForDeviceGone(deviceGlob, 20)
}

type initiatorISCSI struct {
	targetAddr string
	targetPort string
	iqn        string
}

func (iscsi *initiatorISCSI) Connect() (string, error) {
	// iscsiadm -m discovery -t sendtargets -p ip:port
	target := iscsi.targetAddr + ":" + iscsi.targetPort
	cmdLine := []string{"iscsiadm", "-m", "discovery", "-t", "sendtargets", "-p", target}
	err := execWithTimeout(cmdLine, 40)
	if err != nil {
		klog.Errorf("command %v failed: %s", cmdLine, err)
	}
	// iscsiadm -m node -T "iqn" -p ip:port --login
	cmdLine = []string{"iscsiadm", "-m", "node", "-T", iscsi.iqn, "-p", target, "--login"}
	err = execWithTimeout(cmdLine, 40)
	if err != nil {
		klog.Errorf("command %v failed: %s", cmdLine, err)
	}

	deviceGlob := fmt.Sprintf("/dev/disk/by-path/*%s*", iscsi.iqn)
	devicePath, err := waitForDeviceReady(deviceGlob, 20)
	if err != nil {
		return "", err
	}
	return devicePath, nil
}

func (iscsi *initiatorISCSI) Disconnect() error {
	target := iscsi.targetAddr + ":" + iscsi.targetPort
	// iscsiadm -m node -T "iqn" -p ip:port --logout
	cmdLine := []string{"iscsiadm", "-m", "node", "-T", iscsi.iqn, "-p", target, "--logout"}
	err := execWithTimeout(cmdLine, 40)
	if err != nil {
		klog.Errorf("command %v failed: %s", cmdLine, err)
	}

	deviceGlob := fmt.Sprintf("/dev/disk/by-path/*%s*", iscsi.iqn)
	return waitForDeviceGone(deviceGlob, 20)
}

// wait for device file comes up or timeout
func waitForDeviceReady(deviceGlob string, seconds int) (string, error) {
	for i := 0; i <= seconds; i++ {
		time.Sleep(time.Second)
		matches, err := filepath.Glob(deviceGlob)
		if err != nil {
			return "", err
		}
		// two symbol links under /dev/disk/by-id/ to same device
		if len(matches) >= 1 {
			return matches[0], nil
		}
	}
	return "", fmt.Errorf("timed out waiting device ready: %s", deviceGlob)
}

// wait for device file gone or timeout
func waitForDeviceGone(deviceGlob string, seconds int) error {
	for i := 0; i <= seconds; i++ {
		time.Sleep(time.Second)
		matches, err := filepath.Glob(deviceGlob)
		if err != nil {
			return err
		}
		if len(matches) == 0 {
			return nil
		}
	}
	return fmt.Errorf("timed out waiting device gone: %s", deviceGlob)
}

// exec shell command with timeout(in seconds)
func execWithTimeout(cmdLine []string, timeout int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	klog.Infof("running command: %v", cmdLine)
	//nolint:gosec // execWithTimeout assumes valid cmd arguments
	cmd := exec.CommandContext(ctx, cmdLine[0], cmdLine[1:]...)
	output, err := cmd.CombinedOutput()

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return fmt.Errorf("timed out")
	}
	if output != nil {
		klog.Infof("command returned: %s", output)
	}
	return err
}

// SMA initiator implementation
type initiatorSMA struct {
	serverURL     string
	req           *smarpc.CreateDeviceRequest
	volumeContext map[string]string
}

func (sma *initiatorSMA) Connect() (string, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(sma.serverURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		klog.Errorf("failed to connect to SMA grpc server in: %s, %s", sma.serverURL, err)
	}
	client := smarpc.NewStorageManagementAgentClient(conn)

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 42*time.Second)
	defer cancel()

	// Create device
	response, err := client.CreateDevice(ctxTimeout, sma.req)
	if err != nil {
		klog.Errorf("Creating device failed: %s", err)
	}
	klog.Infof("DELME: initiator Connect: CreateDevice response: %+v", response)
	klog.Infof("DELME: volumeID: %s", sma.volumeContext["model"])
	
	volUuid := uuid.MustParse(sma.volumeContext["model"])
	volUuidBytes, err := volUuid.MarshalBinary()
	volUuidB64 := base64.StdEncoding.EncodeToString(volUuidBytes)
	klog.Infof("DELME: volumeID to base64: %s", volUuidB64)
	
	deviceHandle := response.Handle
	attachReq := &smarpc.AttachVolumeRequest{
		Volume:       &smarpc.VolumeParameters{VolumeId: []byte(volUuidB64)},
		DeviceHandle: deviceHandle,
	}

	// Attach volume
	attachRes, err := client.AttachVolume(ctxTimeout, attachReq)
	if err != nil {
		klog.Errorf("Attaching volume failed: %s", err)
	}
	klog.Infof(attachRes.String())

	// Detach volumn
	// echo '{"method": "DetachVolume", "params": {"device_handle": "nvmf-tcp:nqn.2016-06.io.spdk:cnode0", "volume_id": "N553+RGYRTaREBH81jIiwQ=="}}' | ./spdk/scripts/sma-client.py --address 127.0.0.1 --port 5114
	// detachRes, err := client.DetachVolume()
	// if err !=
	/*
		deviceHandle := response.Handle
		volumnID := sma.req.GetNvmfTcp().Subnqn[29:]


		klog.Infof("DELME: deviceHandle %+v", deviceHandle)
		klog.Infof("DELME: Sunnqn %+v", sma.req.GetNvmfTcp().Subnqn)
		klog.Infof("DELME: volumnID %+v", volumnID)

		attachReq := smarpc.AttachVolumeRequest{
			Volume:       &smarpc.VolumeParameters{VolumeId: []byte(volumnID)},
			DeviceHandle: deviceHandle,
		}
		// Attach device
		response, err = client.AttachVolume(ctxTimeout, *attachReq)
		if err != nil {
			klog.Errorf("initiatorSMA.Connect(): not implemented")
		}
		klog.Infof(response.GetHandle())
	*/
	return "", fmt.Errorf("initiatorSMA.Connect(): not implemented error")
}

/*
func (sma *initiatorSMA) Connect() (string, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(sma.serverURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", fmt.Errorf("failed to connect to SMA grpc server in %q: %w", sma.serverURL, err)
	}
	client := smarpc.NewStorageManagementAgentClient(conn)
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 42*time.Second)
	defer cancel()
	response, err := client.CreateDevice(ctxTimeout, sma.req)
	if err != nil {
		klog.Errorf("initiatorSMA.Connect(): not implemented")
	}
	klog.Infof("DELME: initiator Connect: CreateDevice response: %+v", response)

	// TODO: send CreateDevice to SMA
	// TODO: wait for /dev/... to appear.
	return "", fmt.Errorf("initiatorSMA.Connect(): not implemented error")
}
*/

func (sma *initiatorSMA) Disconnect() error {
	klog.Errorf("initiatorSMA.Disconnect(): not implemented")
	return fmt.Errorf("initiatorSMA.Disconnect(): not implemented error")
}

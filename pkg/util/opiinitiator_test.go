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
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	opiapiCommon "github.com/opiproject/opi-api/common/v1/gen/go"
	opiapiStorage "github.com/opiproject/opi-api/storage/v1alpha1/gen/go"
)

// MockNVMfRemoteControllerServiceClient is a mock implementation of the NVMfRemoteControllerServiceClient interface
type MockNVMfRemoteControllerServiceClient struct {
	mock.Mock
}

func (m *MockNVMfRemoteControllerServiceClient) CreateNVMfRemoteController(ctx context.Context, in *opiapiStorage.CreateNVMfRemoteControllerRequest, opts ...grpc.CallOption) (*opiapiStorage.NVMfRemoteController, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*opiapiStorage.NVMfRemoteController), args.Error(1)
}

func (m *MockNVMfRemoteControllerServiceClient) DeleteNVMfRemoteController(ctx context.Context, in *opiapiStorage.DeleteNVMfRemoteControllerRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockNVMfRemoteControllerServiceClient) UpdateNVMfRemoteController(ctx context.Context, in *opiapiStorage.UpdateNVMfRemoteControllerRequest, opts ...grpc.CallOption) (*opiapiStorage.NVMfRemoteController, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*opiapiStorage.NVMfRemoteController), args.Error(1)
}

func (m *MockNVMfRemoteControllerServiceClient) ListNVMfRemoteControllers(ctx context.Context, in *opiapiStorage.ListNVMfRemoteControllersRequest, opts ...grpc.CallOption) (*opiapiStorage.ListNVMfRemoteControllersResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*opiapiStorage.ListNVMfRemoteControllersResponse), args.Error(1)
}

func (m *MockNVMfRemoteControllerServiceClient) GetNVMfRemoteController(ctx context.Context, in *opiapiStorage.GetNVMfRemoteControllerRequest, opts ...grpc.CallOption) (*opiapiStorage.NVMfRemoteController, error) {
	args := m.Called(ctx, in)
	return nil, args.Error(1)
}

func (m *MockNVMfRemoteControllerServiceClient) NVMfRemoteControllerReset(ctx context.Context, in *opiapiStorage.NVMfRemoteControllerResetRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockNVMfRemoteControllerServiceClient) NVMfRemoteControllerStats(ctx context.Context, in *opiapiStorage.NVMfRemoteControllerStatsRequest, opts ...grpc.CallOption) (*opiapiStorage.NVMfRemoteControllerStatsResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*opiapiStorage.NVMfRemoteControllerStatsResponse), args.Error(1)
}

type MockFrontendVirtioBlkServiceClient struct {
	mock.Mock
}

func (i *MockFrontendVirtioBlkServiceClient) CreateVirtioBlk(ctx context.Context, in *opiapiStorage.CreateVirtioBlkRequest, opts ...grpc.CallOption) (*opiapiStorage.VirtioBlk, error) {
	args := i.Called(ctx, in)
	if args.Get(0) != nil {
		return args.Get(0).(*opiapiStorage.VirtioBlk), nil
	} else {
		return nil, args.Error(1)
	}
}
func (i *MockFrontendVirtioBlkServiceClient) DeleteVirtioBlk(ctx context.Context, in *opiapiStorage.DeleteVirtioBlkRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := i.Called(ctx, in)
	if args.Get(0) != nil {
		return args.Get(0).(*emptypb.Empty), nil
	} else {
		return nil, args.Error(1)
	}
}
func (i *MockFrontendVirtioBlkServiceClient) UpdateVirtioBlk(ctx context.Context, in *opiapiStorage.UpdateVirtioBlkRequest, opts ...grpc.CallOption) (*opiapiStorage.VirtioBlk, error) {
	args := i.Called(ctx, in)
	return args.Get(0).(*opiapiStorage.VirtioBlk), args.Error(1)
}
func (i *MockFrontendVirtioBlkServiceClient) ListVirtioBlks(ctx context.Context, in *opiapiStorage.ListVirtioBlksRequest, opts ...grpc.CallOption) (*opiapiStorage.ListVirtioBlksResponse, error) {
	args := i.Called(ctx, in)
	return args.Get(0).(*opiapiStorage.ListVirtioBlksResponse), args.Error(1)
}
func (i *MockFrontendVirtioBlkServiceClient) GetVirtioBlk(ctx context.Context, in *opiapiStorage.GetVirtioBlkRequest, opts ...grpc.CallOption) (*opiapiStorage.VirtioBlk, error) {
	args := i.Called(ctx, in)
	if args.Get(0) != nil {
		return args.Get(0).(*opiapiStorage.VirtioBlk), nil
	} else {
		return nil, args.Error(1)
	}
}
func (i *MockFrontendVirtioBlkServiceClient) VirtioBlkStats(ctx context.Context, in *opiapiStorage.VirtioBlkStatsRequest, opts ...grpc.CallOption) (*opiapiStorage.VirtioBlkStatsResponse, error) {
	args := i.Called(ctx, in)
	return args.Get(0).(*opiapiStorage.VirtioBlkStatsResponse), args.Error(1)
}

type MockFrontendNvmeServiceClient struct {
	mock.Mock
}

func (i *MockFrontendNvmeServiceClient) CreateNVMeSubsystem(ctx context.Context, in *opiapiStorage.CreateNVMeSubsystemRequest, opts ...grpc.CallOption) (*opiapiStorage.NVMeSubsystem, error) {
	args := i.Called(ctx, in)
	if args.Get(0) != nil {
		return args.Get(0).(*opiapiStorage.NVMeSubsystem), nil
	} else {
		return nil, args.Error(1)
	}
}

func (i *MockFrontendNvmeServiceClient) DeleteNVMeSubsystem(ctx context.Context, in *opiapiStorage.DeleteNVMeSubsystemRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := i.Called(ctx, in)
	if args.Get(0) != nil {
		return args.Get(0).(*emptypb.Empty), nil
	} else {
		return nil, args.Error(1)
	}
}
func (i *MockFrontendNvmeServiceClient) UpdateNVMeSubsystem(ctx context.Context, in *opiapiStorage.UpdateNVMeSubsystemRequest, opts ...grpc.CallOption) (*opiapiStorage.NVMeSubsystem, error) {
	return nil, nil
}

func (i *MockFrontendNvmeServiceClient) ListNVMeSubsystems(ctx context.Context, in *opiapiStorage.ListNVMeSubsystemsRequest, opts ...grpc.CallOption) (*opiapiStorage.ListNVMeSubsystemsResponse, error) {
	return nil, nil
}
func (i *MockFrontendNvmeServiceClient) GetNVMeSubsystem(ctx context.Context, in *opiapiStorage.GetNVMeSubsystemRequest, opts ...grpc.CallOption) (*opiapiStorage.NVMeSubsystem, error) {
	args := i.Called(ctx, in)
	if args.Get(0) != nil {
		return args.Get(0).(*opiapiStorage.NVMeSubsystem), nil
	} else {
		return nil, args.Error(1)
	}
}

func (i *MockFrontendNvmeServiceClient) NVMeSubsystemStats(ctx context.Context, in *opiapiStorage.NVMeSubsystemStatsRequest, opts ...grpc.CallOption) (*opiapiStorage.NVMeSubsystemStatsResponse, error) {
	return nil, nil
}

func (i *MockFrontendNvmeServiceClient) CreateNVMeController(ctx context.Context, in *opiapiStorage.CreateNVMeControllerRequest, opts ...grpc.CallOption) (*opiapiStorage.NVMeController, error) {
	args := i.Called(ctx, in)
	if args.Get(0) != nil {
		return args.Get(0).(*opiapiStorage.NVMeController), nil
	} else {
		return nil, args.Error(1)
	}
}

func (i *MockFrontendNvmeServiceClient) DeleteNVMeController(ctx context.Context, in *opiapiStorage.DeleteNVMeControllerRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := i.Called(ctx, in)
	if args.Get(0) != nil {
		return args.Get(0).(*emptypb.Empty), nil
	} else {
		return nil, args.Error(1)
	}
}

func (i *MockFrontendNvmeServiceClient) UpdateNVMeController(ctx context.Context, in *opiapiStorage.UpdateNVMeControllerRequest, opts ...grpc.CallOption) (*opiapiStorage.NVMeController, error) {
	return nil, nil
}

func (i *MockFrontendNvmeServiceClient) ListNVMeControllers(ctx context.Context, in *opiapiStorage.ListNVMeControllersRequest, opts ...grpc.CallOption) (*opiapiStorage.ListNVMeControllersResponse, error) {
	return nil, nil
}

func (i *MockFrontendNvmeServiceClient) GetNVMeController(ctx context.Context, in *opiapiStorage.GetNVMeControllerRequest, opts ...grpc.CallOption) (*opiapiStorage.NVMeController, error) {
	args := i.Called(ctx, in)
	if args.Get(0) != nil {
		return args.Get(0).(*opiapiStorage.NVMeController), nil
	} else {
		return nil, args.Error(1)
	}
}

func (i *MockFrontendNvmeServiceClient) NVMeControllerStats(ctx context.Context, in *opiapiStorage.NVMeControllerStatsRequest, opts ...grpc.CallOption) (*opiapiStorage.NVMeControllerStatsResponse, error) {
	return nil, nil
}
func (i *MockFrontendNvmeServiceClient) CreateNVMeNamespace(ctx context.Context, in *opiapiStorage.CreateNVMeNamespaceRequest, opts ...grpc.CallOption) (*opiapiStorage.NVMeNamespace, error) {
	args := i.Called(ctx, in)
	if args.Get(0) != nil {
		return args.Get(0).(*opiapiStorage.NVMeNamespace), nil
	} else {
		return nil, args.Error(1)
	}
}
func (i *MockFrontendNvmeServiceClient) DeleteNVMeNamespace(ctx context.Context, in *opiapiStorage.DeleteNVMeNamespaceRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := i.Called(ctx, in)
	if args.Get(0) != nil {
		return args.Get(0).(*emptypb.Empty), nil
	} else {
		return nil, args.Error(1)
	}
}
func (i *MockFrontendNvmeServiceClient) UpdateNVMeNamespace(ctx context.Context, in *opiapiStorage.UpdateNVMeNamespaceRequest, opts ...grpc.CallOption) (*opiapiStorage.NVMeNamespace, error) {
	return nil, nil
}
func (i *MockFrontendNvmeServiceClient) ListNVMeNamespaces(ctx context.Context, in *opiapiStorage.ListNVMeNamespacesRequest, opts ...grpc.CallOption) (*opiapiStorage.ListNVMeNamespacesResponse, error) {
	return nil, nil
}
func (i *MockFrontendNvmeServiceClient) GetNVMeNamespace(ctx context.Context, in *opiapiStorage.GetNVMeNamespaceRequest, opts ...grpc.CallOption) (*opiapiStorage.NVMeNamespace, error) {
	args := i.Called(ctx, in)
	if args.Get(0) != nil {
		return args.Get(0).(*opiapiStorage.NVMeNamespace), nil
	} else {
		return nil, args.Error(1)
	}
}
func (i *MockFrontendNvmeServiceClient) NVMeNamespaceStats(ctx context.Context, in *opiapiStorage.NVMeNamespaceStatsRequest, opts ...grpc.CallOption) (*opiapiStorage.NVMeNamespaceStatsResponse, error) {
	return nil, nil
}

func TestCreateNvmfRemoteController(t *testing.T) {
	// Create a mock NVMfRemoteControllerServiceClient
	mockClient := new(MockNVMfRemoteControllerServiceClient)

	// Create an instance of opiCommon
	opi := &opiCommon{
		volumeContext:              map[string]string{},
		timeout:                    time.Second,
		kvmPciBridges:              0,
		opiClient:                  nil,
		nvmfRemoteControllerClient: mockClient,
		nvmfRemoteControllerID:     "",
		devicePath:                 "",
	}

	// Mock the CreateNVMfRemoteController function to return a response with the specified ID
	mockClient.On("CreateNVMfRemoteController", mock.Anything, mock.Anything).Return(&opiapiStorage.NVMfRemoteController{
		Id: &opiapiCommon.ObjectKey{
			Value: "controllerID",
		},
	}, nil)

	// Mock the GetNVMfRemoteController function to return a nil response and an error
	mockClient.On("GetNVMfRemoteController", mock.Anything, mock.Anything).Return(nil, errors.New("Controller does not exist"))

	// Set the necessary volume context values
	opi.volumeContext["targetPort"] = "1234"
	opi.volumeContext["targetAddr"] = "192.168.0.1"
	opi.volumeContext["nqn"] = "nqn-value"
	opi.volumeContext["model"] = "model-value"

	// Call the function under test
	err := opi.createNvmfRemoteController("controllerID")

	// Assert that the GetNVMfRemoteController and CreateNVMfRemoteController functions were called with the expected arguments
	mockClient.AssertCalled(t, "GetNVMfRemoteController", mock.Anything, &opiapiStorage.GetNVMfRemoteControllerRequest{Name: "controllerID"})
	mockClient.AssertCalled(t, "CreateNVMfRemoteController", mock.Anything, &opiapiStorage.CreateNVMfRemoteControllerRequest{
		NvMfRemoteController: &opiapiStorage.NVMfRemoteController{
			Id: &opiapiCommon.ObjectKey{
				Value: "controllerID",
			},
			Trtype:  4,
			Adrfam:  1,
			Traddr:  "192.168.0.1",
			Trsvcid: 1234,
			Subnqn:  "nqn-value",
			Hostnqn: "nqn.2023-04.io.spdk.csi:remote.controller:uuid:" + opi.volumeContext["model"],
		},
	})

	// Assert that the error returned is nil
	assert.NoError(t, err)
}

func TestDeleteNvmfRemoteController_Success(t *testing.T) {
	// Create a mock client
	mockClient := new(MockNVMfRemoteControllerServiceClient)

	// Create the OPI instance with the mock client
	opi := &opiCommon{
		nvmfRemoteControllerClient: mockClient,
	}

	// Set up expectations
	nvmfRemoteControllerID := "controllerID"
	deleteReq := &opiapiStorage.DeleteNVMfRemoteControllerRequest{
		Name:         nvmfRemoteControllerID,
		AllowMissing: true,
	}
	mockClient.On("DeleteNVMfRemoteController", mock.Anything, deleteReq).Return(&emptypb.Empty{}, nil)

	// Call the method under test
	err := opi.deleteNvmfRemoteController(nvmfRemoteControllerID)

	// Assert that the expected methods were called
	mockClient.AssertExpectations(t)

	// Assert that no error was returned
	assert.NoError(t, err)
}

func TestDeleteNvmfRemoteController_Error(t *testing.T) {
	// Create a mock client
	mockClient := new(MockNVMfRemoteControllerServiceClient)

	// Create the OPI instance with the mock client
	opi := &opiCommon{
		nvmfRemoteControllerClient: mockClient,
	}

	// Set up expectations
	nvmfRemoteControllerID := "controllerID"
	deleteReq := &opiapiStorage.DeleteNVMfRemoteControllerRequest{
		Name:         nvmfRemoteControllerID,
		AllowMissing: true,
	}
	expectedErr := errors.New("delete error")
	mockClient.On("DeleteNVMfRemoteController", mock.Anything, deleteReq).Return(&emptypb.Empty{}, expectedErr)

	// Call the method under test
	err := opi.deleteNvmfRemoteController(nvmfRemoteControllerID)

	// Assert that the expected methods were called
	mockClient.AssertExpectations(t)

	// Assert that the correct error was returned
	assert.EqualError(t, err, fmt.Sprintf("OPI.DeleteNVMfRemoteController error: %v", expectedErr))
}

func TestCreateVirtioBlk_Failure(t *testing.T) {
	// Create a mock client
	mockClient := new(MockFrontendVirtioBlkServiceClient)

	// Create a mock NVMfRemoteControllerServiceClient
	mockClient2 := new(MockNVMfRemoteControllerServiceClient)
	// Create the OPI instance with the mock client
	opi := &opiInitiatorVirtioBlk{
		frontendVirtioBlkClient: mockClient,
		// Create an instance of opiCommon
		opi: &opiCommon{
			volumeContext:              map[string]string{},
			timeout:                    time.Second,
			kvmPciBridges:              1,
			opiClient:                  nil,
			nvmfRemoteControllerClient: mockClient2,
			nvmfRemoteControllerID:     "",
			devicePath:                 "",
		},
	}

	// Mock the CreateNVMfRemoteController function to return a response with the specified ID
	mockClient.On("GetVirtioBlk", mock.Anything, mock.Anything).Return(nil, errors.New("Could not find Controller"))

	// Mock the GetNVMfRemoteController function to return a nil response and an error
	mockClient.On("CreateVirtioBlk", mock.Anything, mock.Anything).Return(nil, errors.New("Controller does not exist"))

	bdf, err := opi.createVirtioBlk()
	assert.Equal(t, bdf, "")
	assert.NotEqual(t, err, nil)
}

func TestCreateVirtioBlk_Success(t *testing.T) {
	// Create a mock client
	mockClient := new(MockFrontendVirtioBlkServiceClient)

	// Create a mock NVMfRemoteControllerServiceClient
	mockClient2 := new(MockNVMfRemoteControllerServiceClient)
	// Create the OPI instance with the mock client
	opi := &opiInitiatorVirtioBlk{
		frontendVirtioBlkClient: mockClient,
		// Create an instance of opiCommon
		opi: &opiCommon{
			volumeContext:              map[string]string{},
			timeout:                    time.Second,
			kvmPciBridges:              1,
			opiClient:                  nil,
			nvmfRemoteControllerClient: mockClient2,
			nvmfRemoteControllerID:     "",
			devicePath:                 "",
		},
	}

	// Mock the CreateNVMfRemoteController function to return a response with the specified ID
	mockClient.On("GetVirtioBlk", mock.Anything, mock.Anything).Return(&opiapiStorage.VirtioBlk{}, errors.New("Could not find Controller"))

	// Mock the GetNVMfRemoteController function to return a nil response and an error
	mockClient.On("CreateVirtioBlk", mock.Anything, mock.Anything).Return(&opiapiStorage.VirtioBlk{Id: &opiapiCommon.ObjectKey{Value: "ss"}}, nil)

	bdf, err := opi.createVirtioBlk()
	assert.Equal(t, bdf, "")
	assert.NotEqual(t, err, nil)
}

func TestDeleteVirtioBlk_Success(t *testing.T) {
	// Create a mock client
	mockClient := new(MockFrontendVirtioBlkServiceClient)

	// Create a mock NVMfRemoteControllerServiceClient
	mockClient2 := new(MockNVMfRemoteControllerServiceClient)
	// Create the OPI instance with the mock client
	opi := &opiInitiatorVirtioBlk{
		frontendVirtioBlkClient: mockClient,
		// Create an instance of opiCommon
		opi: &opiCommon{
			volumeContext:              map[string]string{},
			timeout:                    time.Second,
			kvmPciBridges:              1,
			opiClient:                  nil,
			nvmfRemoteControllerClient: mockClient2,
			nvmfRemoteControllerID:     "",
			devicePath:                 "",
		},
	}

	// Mock the CreateNVMfRemoteController function to return a response with the specified ID
	mockClient.On("DeleteVirtioBlk", mock.Anything, mock.Anything).Return(nil, errors.New("Could not find Controller"))

	err := opi.deleteVirtioBlk("virtioBlkID")
	assert.NotEqual(t, err, nil)
}

func TestDeleteVirtioBlk_Failure(t *testing.T) {
	// Create a mock client
	mockClient := new(MockFrontendVirtioBlkServiceClient)

	// Create a mock NVMfRemoteControllerServiceClient
	mockClient2 := new(MockNVMfRemoteControllerServiceClient)
	// Create the OPI instance with the mock client
	opi := &opiInitiatorVirtioBlk{
		frontendVirtioBlkClient: mockClient,
		// Create an instance of opiCommon
		opi: &opiCommon{
			volumeContext:              map[string]string{},
			timeout:                    time.Second,
			kvmPciBridges:              1,
			opiClient:                  nil,
			nvmfRemoteControllerClient: mockClient2,
			nvmfRemoteControllerID:     "",
			devicePath:                 "",
		},
	}

	// Mock the CreateNVMfRemoteController function to return a response with the specified ID
	mockClient.On("DeleteVirtioBlk", mock.Anything, mock.Anything).Return(nil, nil)

	err := opi.deleteVirtioBlk("virtioBlkID")
	assert.Equal(t, err, nil)
}

func TestCreateNVMeSubsystem(t *testing.T) {

	mockClient := new(MockFrontendNvmeServiceClient)
	mockClient2 := new(MockNVMfRemoteControllerServiceClient)
	opi := &opiInitiatorNvme{
		frontendNvmeClient: mockClient,
		// Create an instance of opiCommon
		opi: &opiCommon{
			volumeContext:              map[string]string{},
			timeout:                    time.Second,
			kvmPciBridges:              1,
			opiClient:                  nil,
			nvmfRemoteControllerClient: mockClient2,
			nvmfRemoteControllerID:     "",
			devicePath:                 "",
		},
	}
	mockClient.On("GetNVMeSubsystem", mock.Anything, mock.Anything).Return(nil, errors.New("unable to find key"))
	mockClient.On("CreateNVMeSubsystem", mock.Anything, mock.Anything).Return(&opiapiStorage.NVMeSubsystem{Spec: &opiapiStorage.NVMeSubsystemSpec{Id: &opiapiCommon.ObjectKey{Value: "sub"}}}, nil)

	err := opi.createNVMeSubsystem("sub")
	assert.Equal(t, err, nil)

}

func TestDeleteNVMeSubsystem(t *testing.T) {

	mockClient := new(MockFrontendNvmeServiceClient)
	mockClient2 := new(MockNVMfRemoteControllerServiceClient)
	opi := &opiInitiatorNvme{
		frontendNvmeClient: mockClient,
		// Create an instance of opiCommon
		opi: &opiCommon{
			volumeContext:              map[string]string{},
			timeout:                    time.Second,
			kvmPciBridges:              1,
			opiClient:                  nil,
			nvmfRemoteControllerClient: mockClient2,
			nvmfRemoteControllerID:     "",
			devicePath:                 "",
		},
	}
	mockClient.On("DeleteNVMeSubsystem", mock.Anything, mock.Anything).Return(nil, errors.New("unable to find key"))
	err := opi.deleteNVMeSubsystem("sub")
	assert.NotEqual(t, err, nil)
}

func TestCreateNVMeController(t *testing.T) {

	mockClient := new(MockFrontendNvmeServiceClient)
	mockClient2 := new(MockNVMfRemoteControllerServiceClient)
	opi := &opiInitiatorNvme{
		frontendNvmeClient: mockClient,
		// Create an instance of opiCommon
		opi: &opiCommon{
			volumeContext:              map[string]string{},
			timeout:                    time.Second,
			kvmPciBridges:              1,
			opiClient:                  nil,
			nvmfRemoteControllerClient: mockClient2,
			nvmfRemoteControllerID:     "",
			devicePath:                 "",
		},
	}
	mockClient.On("GetNVMeController", mock.Anything, mock.Anything).Return(nil, errors.New("unable to find key"))
	mockClient.On("CreateNVMeController", mock.Anything, mock.Anything).Return(&opiapiStorage.NVMeController{Spec: &opiapiStorage.NVMeControllerSpec{Id: &opiapiCommon.ObjectKey{Value: "sub"}}}, nil)
	_, err := opi.createNVMeController()
	assert.Equal(t, err, nil)
}

func TestDeleteNVMeController(t *testing.T) {

	mockClient := new(MockFrontendNvmeServiceClient)
	mockClient2 := new(MockNVMfRemoteControllerServiceClient)
	opi := &opiInitiatorNvme{
		frontendNvmeClient: mockClient,
		// Create an instance of opiCommon
		opi: &opiCommon{
			volumeContext:              map[string]string{},
			timeout:                    time.Second,
			kvmPciBridges:              1,
			opiClient:                  nil,
			nvmfRemoteControllerClient: mockClient2,
			nvmfRemoteControllerID:     "",
			devicePath:                 "",
		},
	}
	mockClient.On("DeleteNVMeController", mock.Anything, mock.Anything).Return(nil, errors.New("unable to find key"))
	err := opi.deleteNVMeController("sub")
	assert.NotEqual(t, err, nil)
}

func TestCreateNVMeNamespace(t *testing.T) {

	mockClient := new(MockFrontendNvmeServiceClient)
	mockClient2 := new(MockNVMfRemoteControllerServiceClient)
	opi := &opiInitiatorNvme{
		frontendNvmeClient: mockClient,
		// Create an instance of opiCommon
		opi: &opiCommon{
			volumeContext:              map[string]string{},
			timeout:                    time.Second,
			kvmPciBridges:              1,
			opiClient:                  nil,
			nvmfRemoteControllerClient: mockClient2,
			nvmfRemoteControllerID:     "",
			devicePath:                 "",
		},
	}
	mockClient.On("GetNVMeNamespace", mock.Anything, mock.Anything).Return(nil, errors.New("unable to find key"))
	mockClient.On("CreateNVMeNamespace", mock.Anything, mock.Anything).Return(nil, errors.New("unable to find key"))
	err := opi.createNVMeNamespace("sub")
	assert.NotEqual(t, err, nil)
}

func TestDeleteNVMeNamespace(t *testing.T) {

	mockClient := new(MockFrontendNvmeServiceClient)
	mockClient2 := new(MockNVMfRemoteControllerServiceClient)
	opi := &opiInitiatorNvme{
		frontendNvmeClient: mockClient,
		// Create an instance of opiCommon
		opi: &opiCommon{
			volumeContext:              map[string]string{},
			timeout:                    time.Second,
			kvmPciBridges:              1,
			opiClient:                  nil,
			nvmfRemoteControllerClient: mockClient2,
			nvmfRemoteControllerID:     "",
			devicePath:                 "",
		},
	}
	mockClient.On("DeleteNVMeNamespace", mock.Anything, mock.Anything).Return(nil, errors.New("unable to find key"))
	err := opi.deleteNVMeNamespace("sub")
	assert.NotEqual(t, err, nil)
}

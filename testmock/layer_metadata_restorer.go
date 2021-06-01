// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/buildpacks/lifecycle (interfaces: LayerMetadataRestorer)

// Package testmock is a generated GoMock package.
package testmock

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"

	lifecycle "github.com/buildpacks/lifecycle"
	buildpack "github.com/buildpacks/lifecycle/buildpack"
	platform "github.com/buildpacks/lifecycle/platform"
)

// MockLayerMetadataRestorer is a mock of LayerMetadataRestorer interface.
type MockLayerMetadataRestorer struct {
	ctrl     *gomock.Controller
	recorder *MockLayerMetadataRestorerMockRecorder
}

// MockLayerMetadataRestorerMockRecorder is the mock recorder for MockLayerMetadataRestorer.
type MockLayerMetadataRestorerMockRecorder struct {
	mock *MockLayerMetadataRestorer
}

// NewMockLayerMetadataRestorer creates a new mock instance.
func NewMockLayerMetadataRestorer(ctrl *gomock.Controller) *MockLayerMetadataRestorer {
	mock := &MockLayerMetadataRestorer{ctrl: ctrl}
	mock.recorder = &MockLayerMetadataRestorerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLayerMetadataRestorer) EXPECT() *MockLayerMetadataRestorerMockRecorder {
	return m.recorder
}

// CacheIsValid mocks base method.
func (m *MockLayerMetadataRestorer) CacheIsValid(arg0, arg1 string, arg2 lifecycle.BpLayer) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CacheIsValid", arg0, arg1, arg2)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CacheIsValid indicates an expected call of CacheIsValid.
func (mr *MockLayerMetadataRestorerMockRecorder) CacheIsValid(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CacheIsValid", reflect.TypeOf((*MockLayerMetadataRestorer)(nil).CacheIsValid), arg0, arg1, arg2)
}

// Restore mocks base method.
func (m *MockLayerMetadataRestorer) Restore(arg0 []buildpack.GroupBuildpack, arg1 platform.LayersMetadata, arg2 platform.CacheMetadata) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Restore", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Restore indicates an expected call of Restore.
func (mr *MockLayerMetadataRestorerMockRecorder) Restore(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Restore", reflect.TypeOf((*MockLayerMetadataRestorer)(nil).Restore), arg0, arg1, arg2)
}

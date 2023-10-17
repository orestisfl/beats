// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/elastic/beats/v8/libbeat/beat (interfaces: Client,Pipeline)

// Package awss3 is a generated GoMock package.
package awss3

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"

	beat "github.com/elastic/beats/v8/libbeat/beat"
)

// MockBeatClient is a mock of Client interface.
type MockBeatClient struct {
	ctrl     *gomock.Controller
	recorder *MockBeatClientMockRecorder
}

// MockBeatClientMockRecorder is the mock recorder for MockBeatClient.
type MockBeatClientMockRecorder struct {
	mock *MockBeatClient
}

// NewMockBeatClient creates a new mock instance.
func NewMockBeatClient(ctrl *gomock.Controller) *MockBeatClient {
	mock := &MockBeatClient{ctrl: ctrl}
	mock.recorder = &MockBeatClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBeatClient) EXPECT() *MockBeatClientMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockBeatClient) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockBeatClientMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockBeatClient)(nil).Close))
}

// Publish mocks base method.
func (m *MockBeatClient) Publish(arg0 beat.Event) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Publish", arg0)
}

// Publish indicates an expected call of Publish.
func (mr *MockBeatClientMockRecorder) Publish(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Publish", reflect.TypeOf((*MockBeatClient)(nil).Publish), arg0)
}

// PublishAll mocks base method.
func (m *MockBeatClient) PublishAll(arg0 []beat.Event) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PublishAll", arg0)
}

// PublishAll indicates an expected call of PublishAll.
func (mr *MockBeatClientMockRecorder) PublishAll(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PublishAll", reflect.TypeOf((*MockBeatClient)(nil).PublishAll), arg0)
}

// MockBeatPipeline is a mock of Pipeline interface.
type MockBeatPipeline struct {
	ctrl     *gomock.Controller
	recorder *MockBeatPipelineMockRecorder
}

// MockBeatPipelineMockRecorder is the mock recorder for MockBeatPipeline.
type MockBeatPipelineMockRecorder struct {
	mock *MockBeatPipeline
}

// NewMockBeatPipeline creates a new mock instance.
func NewMockBeatPipeline(ctrl *gomock.Controller) *MockBeatPipeline {
	mock := &MockBeatPipeline{ctrl: ctrl}
	mock.recorder = &MockBeatPipelineMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBeatPipeline) EXPECT() *MockBeatPipelineMockRecorder {
	return m.recorder
}

// Connect mocks base method.
func (m *MockBeatPipeline) Connect() (beat.Client, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Connect")
	ret0, _ := ret[0].(beat.Client)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Connect indicates an expected call of Connect.
func (mr *MockBeatPipelineMockRecorder) Connect() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Connect", reflect.TypeOf((*MockBeatPipeline)(nil).Connect))
}

// ConnectWith mocks base method.
func (m *MockBeatPipeline) ConnectWith(arg0 beat.ClientConfig) (beat.Client, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConnectWith", arg0)
	ret0, _ := ret[0].(beat.Client)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ConnectWith indicates an expected call of ConnectWith.
func (mr *MockBeatPipelineMockRecorder) ConnectWith(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConnectWith", reflect.TypeOf((*MockBeatPipeline)(nil).ConnectWith), arg0)
}

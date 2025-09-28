// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package cbreceiver

import (
	"testing"

	"github.com/elastic/beats/v7/libbeat/otelbeat/oteltest"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component/componentstatus"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewReceiver(t *testing.T) {
	config := Config{
		Beatconfig: map[string]any{
			"cloudbeat": map[string]any{
				"type":   "cloudbeat/cis_k8s",
				"period": "3h",
				"config": map[string]any{
					"v1": map[string]any{
						"benchmark": "cis_k8s",
					},
				},
			},
			"output": map[string]any{
				"otelconsumer": map[string]any{},
			},
			"logging": map[string]any{
				"level": "debug",
				"selectors": []string{
					"*",
				},
			},
			"path.home": t.TempDir(),
		},
	}

	oteltest.CheckReceivers(oteltest.CheckReceiversParams{
		T: t,
		Receivers: []oteltest.ReceiverConfig{
			{
				Name:    "r1",
				Beat:    "cloudbeat",
				Config:  &config,
				Factory: NewFactory(),
			},
		},
		Status: oteltest.ExpectedStatus{
			Status: componentstatus.StatusOK,
		},
		AssertFunc: func(c *assert.CollectT, logs map[string][]mapstr.M, zapLogs *observer.ObservedLogs) {
			// We don't expect any logs since there are no inputs that would generate data
			// in a test environment without a real cloud provider.
			// The main goal is to ensure the receiver can be created, started, and shutdown without errors.
		},
	})
}

// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

//go:build !integration

package mixer

import (
	"testing"

	mbtest "github.com/elastic/beats/v8/metricbeat/mb/testing"

	_ "github.com/elastic/beats/v8/x-pack/metricbeat/module/istio"
)

func TestData(t *testing.T) {
	mbtest.TestDataFiles(t, "istio", "mixer")
}

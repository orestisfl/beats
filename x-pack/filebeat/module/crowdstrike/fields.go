// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

// Code generated by beats/dev-tools/cmd/asset/asset.go - DO NOT EDIT.

package crowdstrike

import (
	"github.com/elastic/beats/v8/libbeat/asset"
)

func init() {
	if err := asset.SetFields("filebeat", "crowdstrike", asset.ModuleFieldsPri, AssetCrowdstrike); err != nil {
		panic(err)
	}
}

// AssetCrowdstrike returns asset data.
// This is the base64 encoded zlib format compressed contents of module/crowdstrike.
func AssetCrowdstrike() string {
	return "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"
}

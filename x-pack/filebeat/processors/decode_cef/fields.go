// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

// Code generated by beats/dev-tools/cmd/asset/asset.go - DO NOT EDIT.

package decode_cef

import (
	"github.com/elastic/beats/v8/libbeat/asset"
)

func init() {
	if err := asset.SetFields("filebeat", "decode_cef", asset.ModuleFieldsPri, AssetDecodeCef); err != nil {
		panic(err)
	}
}

// AssetDecodeCef returns asset data.
// This is the base64 encoded zlib format compressed contents of processors/decode_cef.
func AssetDecodeCef() string {
	return "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"
}

// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

// Code generated by beats/dev-tools/cmd/asset/asset.go - DO NOT EDIT.

package coredns

import (
	"github.com/elastic/beats/v8/libbeat/asset"
)

func init() {
	if err := asset.SetFields("filebeat", "coredns", asset.ModuleFieldsPri, AssetCoredns); err != nil {
		panic(err)
	}
}

// AssetCoredns returns asset data.
// This is the base64 encoded zlib format compressed contents of module/coredns.
func AssetCoredns() string {
	return "eJy8zjFS9DAMBeA+p3gX2P8ALv4GWmg4gTd+znrwWkZWCu/pGbMhZBhoUflGep9OeGV3mEUZSpsAS5bp8LAHgW3WVC1Jcfg/AcCThDUTURQXX0JOZUGWpaGqhHVmwLkfKmNiDs19nJ5Q/JVHcIz1SodFZa1b8oM6ZjvbKuGjUVFErz6nmx/b2+rR/FLfVmr/19KNe+OdTsW4UPc0jkpzOHdj29NfngJGIyTCLsTj88vdmb7hylalNP6B/0lN7wEAAP//HDeN8A=="
}

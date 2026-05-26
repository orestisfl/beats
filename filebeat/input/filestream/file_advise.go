// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package filestream

import "os"

// pageSize is the OS page size, queried once at init. fadvise length
// arguments must be rounded up to a multiple of this when the caller wants
// sub-page ranges to be evicted (see fadviseDontNeed).
var pageSize = int64(os.Getpagesize())

// pageAlignUp rounds n up to the nearest multiple of the OS page size. It
// must be used on the length passed to fadviseDontNeed for sub-page ranges
// the caller wants actually evicted.
func pageAlignUp(n int64) int64 {
	return (n + pageSize - 1) / pageSize * pageSize
}

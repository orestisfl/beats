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

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFadviseDontNeed_NoError verifies that calling fadvise on a regular
// file does not return an error on platforms that implement it, and is a
// no-op on platforms that do not.
func TestFadviseDontNeed_NoError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fadvise.dat")
	// Write enough data so that there's at least one page to drop.
	require.NoError(t, os.WriteFile(path, bytes.Repeat([]byte("x"), 64*1024), 0o600))

	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	// Length 0 means "from offset to end of file" on Linux. Other platforms
	// return nil unconditionally.
	assert.NoError(t, fadviseDontNeed(f, 0, 0), "fadviseDontNeed should not error")
	assert.NoError(t, fadviseDontNeed(f, 0, 4096), "fadviseDontNeed should not error for a specific range")
}

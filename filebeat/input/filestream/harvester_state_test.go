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
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	loginp "github.com/elastic/beats/v7/filebeat/input/filestream/internal/input-logfile"
	"github.com/elastic/beats/v7/libbeat/common/file"
)

func completeDesc(sum string) loginp.FileDescriptor {
	return loginp.FileDescriptor{Fingerprint: loginp.FingerprintID{Sum: sum}}
}

func growingDesc(raw string) loginp.FileDescriptor {
	return loginp.FileDescriptor{Fingerprint: loginp.FingerprintID{Raw: raw}}
}

// nonZeroOSState returns a real, non-zero StateOS by stat-ing a fresh temp
// file. The StateOS layout is platform specific (inode/device vs Windows file
// id), so deriving it from an actual file keeps the tests cross-platform.
// Distinct files yield distinct identities, so callers get different StateOS
// values on each call.
func nonZeroOSState(t *testing.T) file.StateOS {
	t.Helper()
	path := filepath.Join(t.TempDir(), "f")
	require.NoError(t, os.WriteFile(path, []byte("x"), 0o600), "writing temp file for StateOS")
	fi, err := os.Stat(path)
	require.NoError(t, err, "stat temp file for StateOS")
	st := file.GetOSState(fi)
	require.NotEqual(t, file.StateOS{}, st, "stat must produce a non-zero StateOS")
	return st
}

func TestFileStateTable_RegisterPinLookup(t *testing.T) {
	tbl := newFileStateTable()

	// No entry yet.
	_, ok := tbl.LookupOSState("id-1")
	assert.False(t, ok, "LookupOSState must fail before any Register")

	h := tbl.Register("id-1", completeDesc("sum-1"))
	assert.NotNil(t, h, "Register must return a handle")

	// Registered but not pinned: no OS state yet.
	_, ok = tbl.LookupOSState("id-1")
	assert.False(t, ok, "LookupOSState must fail on an unpinned entry")

	// FingerprintSum reflects the descriptor passed to Register.
	assert.Equal(t, "sum-1", h.FingerprintSum(),
		"FingerprintSum must return the completed Sum passed to Register")

	want := nonZeroOSState(t)
	h.PinOSState(want)
	got, ok := tbl.LookupOSState("id-1")
	assert.True(t, ok, "LookupOSState must succeed on a pinned, non-zero entry")
	assert.Equal(t, want, got, "LookupOSState must return the pinned StateOS")
}

func TestFileStateTable_LookupZeroStateOSReadsAsNoPin(t *testing.T) {
	tbl := newFileStateTable()
	h := tbl.Register("id-1", completeDesc("sum-1"))

	// A zero StateOS (e.g. Windows loadFileId failing) must read as "no pin".
	h.PinOSState(file.StateOS{})
	_, ok := tbl.LookupOSState("id-1")
	assert.False(t, ok, "a zero pinned StateOS must read as no pin")
}

func TestFileStateTable_FingerprintSum(t *testing.T) {
	tbl := newFileStateTable()

	complete := tbl.Register("complete", completeDesc("the-sum"))
	assert.Equal(t, "the-sum", complete.FingerprintSum(),
		"a completed fingerprint must expose its Sum")

	growing := tbl.Register("growing", growingDesc("deadbeef"))
	assert.Empty(t, growing.FingerprintSum(),
		"an incomplete (growing) fingerprint must expose no Sum")
}

func TestFileStateTable_UpdateDescriptorUpdatesIfPresent(t *testing.T) {
	tbl := newFileStateTable()

	// No harvester registered: update must not create a ghost entry.
	tbl.UpdateDescriptor("absent", completeDesc("sum"))
	_, ok := tbl.LookupOSState("absent")
	assert.False(t, ok, "UpdateDescriptor must not insert an entry for an absent key")

	h := tbl.Register("id-1", growingDesc("deadbeef"))
	assert.Empty(t, h.FingerprintSum(),
		"handle must start below threshold with no Sum")

	// Threshold crossing: the scanner now reports a completed fingerprint.
	tbl.UpdateDescriptor("id-1", completeDesc("final-sum"))
	assert.Equal(t, "final-sum", h.FingerprintSum(),
		"UpdateDescriptor must make the completed Sum visible on the handle")
}

func TestFileStateTable_RekeyPreservesHandle(t *testing.T) {
	tbl := newFileStateTable()
	h := tbl.Register("old-id", completeDesc("sum-1"))
	pinned := nonZeroOSState(t)
	h.PinOSState(pinned)

	tbl.Rekey("old-id", "new-id")

	// The pin and descriptor travel with the handle to the new key.
	_, ok := tbl.LookupOSState("old-id")
	assert.False(t, ok, "old key must no longer resolve after Rekey")

	got, ok := tbl.LookupOSState("new-id")
	assert.True(t, ok, "new key must resolve to the migrated handle after Rekey")
	assert.Equal(t, pinned, got, "Rekey must preserve the pinned StateOS")

	// The same handle keeps exposing its fingerprint after migration.
	assert.Equal(t, "sum-1", h.FingerprintSum(),
		"Rekey must preserve the handle's descriptor")

	// UpdateDescriptor now lands under the new key.
	tbl.UpdateDescriptor("new-id", completeDesc("sum-2"))
	assert.Equal(t, "sum-2", h.FingerprintSum(),
		"UpdateDescriptor under the new key must reach the migrated handle")
	tbl.UpdateDescriptor("old-id", completeDesc("ignored"))
	assert.Equal(t, "sum-2", h.FingerprintSum(),
		"UpdateDescriptor under the stale old key must not reach the handle")
}

func TestFileStateTable_RekeyNoHandleIsNoOp(t *testing.T) {
	tbl := newFileStateTable()
	tbl.Rekey("absent", "new-id")
	_, ok := tbl.LookupOSState("new-id")
	assert.False(t, ok, "Rekey of an absent key must not create an entry")
}

func TestFileStateTable_DeregisterCompareAndDelete(t *testing.T) {
	t.Run("removes its own entry", func(t *testing.T) {
		tbl := newFileStateTable()
		h := tbl.Register("id-1", completeDesc("sum"))
		h.PinOSState(nonZeroOSState(t))

		tbl.Deregister(h)
		_, ok := tbl.LookupOSState("id-1")
		assert.False(t, ok, "Deregister must remove the handle's own entry")
	})

	t.Run("displaced handle Deregister no-ops under Restart overlap", func(t *testing.T) {
		tbl := newFileStateTable()

		// Restart overlap: the new harvester registers under the same key
		// before the old one deregisters (overwrite-on-insert, newest wins).
		old := tbl.Register("id-1", completeDesc("old"))
		newer := tbl.Register("id-1", completeDesc("new"))
		newerPin := nonZeroOSState(t)
		newer.PinOSState(newerPin)

		// The displaced (old) handle deregisters: it must NOT evict the newer
		// handle that now owns the key.
		tbl.Deregister(old)
		got, ok := tbl.LookupOSState("id-1")
		assert.True(t, ok, "the newer handle must survive the displaced handle's Deregister")
		assert.Equal(t, newerPin, got, "the surviving entry must be the newer handle's")

		// The newer handle can still deregister itself.
		tbl.Deregister(newer)
		_, ok = tbl.LookupOSState("id-1")
		assert.False(t, ok, "the newer handle must be able to deregister itself")
	})

	t.Run("Deregister after Rekey removes at the current key", func(t *testing.T) {
		tbl := newFileStateTable()
		h := tbl.Register("old-id", completeDesc("sum"))
		h.PinOSState(nonZeroOSState(t))
		tbl.Rekey("old-id", "new-id")

		tbl.Deregister(h)
		_, ok := tbl.LookupOSState("new-id")
		assert.False(t, ok, "Deregister must remove the handle at its post-Rekey key")
	})
}

func TestFileStateTable_NilSafety(t *testing.T) {
	var tbl *fileStateTable
	var h *openFileState

	// None of these must panic on a nil table / nil handle.
	assert.Nil(t, tbl.Register("id", completeDesc("sum")),
		"Register on a nil table must return nil")
	tbl.UpdateDescriptor("id", completeDesc("sum"))
	tbl.Rekey("a", "b")
	tbl.Deregister(h)
	_, ok := tbl.LookupOSState("id")
	assert.False(t, ok, "LookupOSState on a nil table must report no pin")

	h.PinOSState(nonZeroOSState(t))
	assert.Empty(t, h.FingerprintSum(), "FingerprintSum on a nil handle must be empty")

	// A handle obtained from a real table must tolerate a nil-table Deregister path too.
	realTable := newFileStateTable()
	realHandle := realTable.Register("id", completeDesc("sum"))
	tbl.Deregister(realHandle) // nil table, real handle: still a no-op, no panic
}

// TestFileStateTable_ConcurrentAccess exercises every writer and reader from
// separate goroutines so the race detector can flag unguarded access. Run with
// -race.
func TestFileStateTable_ConcurrentAccess(t *testing.T) {
	tbl := newFileStateTable()
	pinned := nonZeroOSState(t)

	const workers = 16
	var wg sync.WaitGroup
	wg.Add(workers * 2)

	for i := 0; i < workers; i++ {
		key := "id-" + strconv.Itoa(i)

		// Harvester goroutine: Register, Pin, FingerprintSum, Deregister.
		go func(key string) {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				h := tbl.Register(key, growingDesc("dead"))
				h.PinOSState(pinned)
				_ = h.FingerprintSum()
				tbl.Deregister(h)
			}
		}(key)

		// Prospector goroutine: UpdateDescriptor, Rekey, LookupOSState.
		go func(key string) {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				tbl.UpdateDescriptor(key, completeDesc("sum"))
				_, _ = tbl.LookupOSState(key)
				tbl.Rekey(key, key+"-moved")
				tbl.Rekey(key+"-moved", key)
			}
		}(key)
	}

	wg.Wait()
}

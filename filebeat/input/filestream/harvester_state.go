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
	"sync"
	"sync/atomic"

	loginp "github.com/elastic/beats/v7/filebeat/input/filestream/internal/input-logfile"
	"github.com/elastic/beats/v7/libbeat/common/file"
)

// fileStateTable is the live communication channel between an input's scanner
// (the prospector/fileWatcher) and its harvesters. There is one table per
// input.
//
// It is keyed by Source.Name() — the "<identity>::<value>" tail that both the
// prospector (src.Name()) and the harvester (fs.Name()) compute for the same
// file. The constant "filestream::<inputID>::" registry prefix (manager.go) is
// NOT part of the key; only the identity tail is.
//
// The table lets a harvester expose live scanner state (the latest
// FileDescriptor, hence the completed fingerprint) and lets the prospector
// learn the fstat identity of the file a harvester currently holds open, so it
// can decide whether a running harvester is still reading the backing file the
// identity now lives on.
//
// Concurrency: mu is a strict leaf lock. Writers are the harvester goroutine
// (Register/PinOSState/Deregister) and the prospector event loop
// (UpdateDescriptor/Rekey). Readers are the prospector (LookupOSState) and the
// harvester hot path (openFileState.FingerprintSum, atomic-only, no lock). All
// methods are safe to call on a nil receiver so tests can build prospectors and
// harvesters without a table.
type fileStateTable struct {
	mu      sync.Mutex
	entries map[string]*openFileState
}

// openFileState is the per-open-file handle shared between one harvester and
// the prospector. The harvester creates it on Register and removes it on
// Deregister; the prospector reads and re-keys it while it lives.
type openFileState struct {
	table *fileStateTable

	// name is the CURRENT table key for this handle. It is guarded by
	// table.mu because Rekey mutates it when a growing fingerprint migrates.
	name string

	// os is the fstat identity of the OPEN fd; it is the zero value until the
	// harvester pins it via PinOSState. Guarded by table.mu.
	os file.StateOS
	// pinned reports whether os has been set. Guarded by table.mu.
	pinned bool

	// desc is the latest descriptor the scanner produced for this file. It is
	// read lock-free on the harvester hot path (FingerprintSum), so it is kept
	// behind an atomic pointer rather than under table.mu.
	desc atomic.Pointer[loginp.FileDescriptor]
}

func newFileStateTable() *fileStateTable {
	return &fileStateTable{
		entries: make(map[string]*openFileState),
	}
}

// Register records a newly opened harvester for name and returns its handle.
// It is called only by the harvester goroutine. Insertion overwrites any
// existing entry for name (newest open wins): a displaced handle keeps working
// on its own fd, and its later Deregister no-ops because Deregister
// compare-and-deletes at the handle's current key.
func (t *fileStateTable) Register(name string, desc loginp.FileDescriptor) *openFileState {
	if t == nil {
		return nil
	}

	h := &openFileState{
		table: t,
		name:  name,
	}
	h.desc.Store(&desc)

	t.mu.Lock()
	t.entries[name] = h
	t.mu.Unlock()

	return h
}

// PinOSState records the fstat identity of the harvester's open fd. It is
// called once, by the harvester goroutine, right after the file is opened.
func (h *openFileState) PinOSState(st file.StateOS) {
	if h == nil {
		return
	}

	h.table.mu.Lock()
	h.os = st
	h.pinned = true
	h.table.mu.Unlock()
}

// Deregister removes h from the table, but only if h is still the handle
// registered under its current key. This compare-and-delete mirrors
// reader.unsafeRemove / readerGroup.migrate: it makes Deregister safe both when
// a newer Register (Restart overlap) displaced h and after Rekey moved h to a
// new key.
func (t *fileStateTable) Deregister(h *openFileState) {
	if t == nil || h == nil {
		return
	}

	t.mu.Lock()
	if t.entries[h.name] == h {
		delete(t.entries, h.name)
	}
	t.mu.Unlock()
}

// UpdateDescriptor refreshes the scanner descriptor for name. It is called by
// the prospector event loop. It updates in place and never inserts: only
// harvesters create entries, so an absent name means no open harvester and the
// update is dropped (otherwise every tracked file would grow a ghost entry).
func (t *fileStateTable) UpdateDescriptor(name string, desc loginp.FileDescriptor) {
	if t == nil {
		return
	}

	t.mu.Lock()
	h, ok := t.entries[name]
	t.mu.Unlock()
	if !ok {
		return
	}

	h.desc.Store(&desc)
}

// Rekey moves the handle registered under old to new. It is called by the
// prospector only after a successful registry key migration (UpdateKey), so the
// table key tracks the registry key. It no-ops when no handle exists for old.
func (t *fileStateTable) Rekey(old, new string) {
	if t == nil || old == new {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	h, ok := t.entries[old]
	if !ok {
		return
	}
	delete(t.entries, old)
	h.name = new
	t.entries[new] = h
}

// LookupOSState returns the pinned fstat identity for name. ok is true only
// when a handle exists, has been pinned, AND the pinned StateOS is non-zero. A
// zero StateOS (e.g. Windows loadFileId failing on a locked/delete-pending
// file) must read as "no pin", never as a match against a zero-valued query.
func (t *fileStateTable) LookupOSState(name string) (file.StateOS, bool) {
	if t == nil {
		return file.StateOS{}, false
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	h, ok := t.entries[name]
	if !ok || !h.pinned || h.os == (file.StateOS{}) {
		return file.StateOS{}, false
	}
	return h.os, true
}

// FingerprintSum returns the file's completed SHA-256 fingerprint, or "" when
// the fingerprint is not yet complete or no descriptor is set. It is a lock-free
// atomic load on the harvester hot path. Only the completed Sum is exposed; the
// raw hex header (possible PII) never leaves the table.
func (h *openFileState) FingerprintSum() string {
	if h == nil {
		return ""
	}

	d := h.desc.Load()
	if d == nil || !d.Fingerprint.Complete() {
		return ""
	}
	return d.Fingerprint.Sum
}

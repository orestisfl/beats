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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDirCacheSequentialReaderRefreshes(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.log"), nil, 0o600), "failed to create initial file")

	c := newDirCache()
	names, sequence, err := c.readDirNames(dir, 0)
	require.NoError(t, err, "failed to warm directory cache")
	assert.Equal(t, []string{"a.log"}, names, "initial listing differs")
	assert.NotZero(t, sequence, "an OS read must assign a sequence")

	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.log"), nil, 0o600), "failed to create second file")

	names, nextSequence, err := c.readDirNames(dir, sequence)
	require.NoError(t, err, "failed to refresh directory cache")
	assert.Equal(t, []string{"a.log", "b.log"}, names, "a sequential read must refresh its own listing")
	assert.Greater(t, nextSequence, sequence, "a refresh must advance the sequence")
}

func TestDirCacheDifferentReaderSharesRefresh(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.log"), nil, 0o600), "failed to create initial file")

	c := newDirCache()
	firstNames, firstSequence, err := c.readDirNames(dir, 0)
	require.NoError(t, err, "first reader failed")
	require.Equal(t, []string{"a.log"}, firstNames, "first listing differs")

	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.log"), nil, 0o600), "failed to create second file")

	secondNames, secondSequence, err := c.readDirNames(dir, 0)
	require.NoError(t, err, "second reader failed")
	assert.Equal(t, []string{"a.log"}, secondNames, "a different reader should share the fresh cached listing")
	assert.Equal(t, firstSequence, secondSequence, "a cache hit must retain the listing sequence")

	firstNames, refreshedSequence, err := c.readDirNames(dir, firstSequence)
	require.NoError(t, err, "first reader failed to refresh")
	require.Equal(t, []string{"a.log", "b.log"}, firstNames, "first reader did not refresh its own listing")
	require.Greater(t, refreshedSequence, firstSequence, "refresh did not advance the sequence")

	secondNames, secondSequence, err = c.readDirNames(dir, secondSequence)
	require.NoError(t, err, "second reader failed to consume refreshed listing")
	assert.Equal(t, []string{"a.log", "b.log"}, secondNames, "second reader did not receive the refreshed listing")
	assert.Equal(t, refreshedSequence, secondSequence, "second reader should reuse the refreshed sequence")
}

func TestDirCacheAbsoluteMaxAge(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.log"), nil, 0o600), "failed to create initial file")

	c := newDirCache()
	_, sequence, err := c.readDirNames(dir, 0)
	require.NoError(t, err, "failed to warm directory cache")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.log"), nil, 0o600), "failed to create second file")

	// Even a reader that has not seen this sequence must not consume a listing
	// older than the cache's absolute age bound.
	c.entry(dir).fetched.Store(time.Now().Add(-maxDirCacheAge).UnixNano())
	names, nextSequence, err := c.readDirNames(dir, 0)
	require.NoError(t, err, "failed to refresh expired listing")
	assert.Equal(t, []string{"a.log", "b.log"}, names, "expired listing was reused")
	assert.Greater(t, nextSequence, sequence, "refreshing an expired listing must advance the sequence")
}

func TestDirCacheNamesAndEntriesShareSequence(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.log"), nil, 0o600), "failed to create initial file")

	c := newDirCache()
	_, namesSequence, err := c.readDirNames(dir, 0)
	require.NoError(t, err, "names read failed")

	// Names cannot answer a request for entry types, so this requires another OS read.
	entries, entriesSequence, err := c.readDirEntries(dir, 0)
	require.NoError(t, err, "entries read failed")
	require.Len(t, entries, 1, "entry listing differs")
	require.Greater(t, entriesSequence, namesSequence, "entries read must advance the sequence")

	// Names can be derived from the newer entry snapshot without another OS read.
	names, derivedSequence, err := c.readDirNames(dir, namesSequence)
	require.NoError(t, err, "derived names read failed")
	assert.Equal(t, []string{"a.log"}, names, "derived names differ")
	assert.Equal(t, entriesSequence, derivedSequence, "derived names must retain the entries sequence")
}

func TestDirCacheErrorNotCached(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing")
	c := newDirCache()

	_, sequence, err := c.readDirNames(dir, 0)
	require.Error(t, err, "reading a missing directory must fail")
	assert.Zero(t, sequence, "a failed read must not return a sequence")
	assert.Zero(t, c.entry(dir).fetched.Load(), "a failed read must invalidate the entry")

	require.NoError(t, os.Mkdir(dir, 0o700), "failed to create directory")
	names, sequence, err := c.readDirNames(dir, 0)
	require.NoError(t, err, "the read after creating the directory must retry")
	assert.Empty(t, names, "new directory should be empty")
	assert.NotZero(t, sequence, "successful retry must assign a sequence")
}

func TestDirCacheConcurrentReaders(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.log"), nil, 0o600), "failed to create initial file")

	c := newDirCache()
	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	results := make([][]string, goroutines)
	sequences := make([]dirCacheSequence, goroutines)
	errs := make([]error, goroutines)
	for i := range goroutines {
		go func(i int) {
			defer wg.Done()
			results[i], sequences[i], errs[i] = c.readDirNames(dir, 0)
		}(i)
	}
	wg.Wait()

	for i := range goroutines {
		require.NoError(t, errs[i], "reader %d failed", i)
		assert.Equal(t, []string{"a.log"}, results[i], "reader %d returned a different listing", i)
		assert.Equal(t, sequences[0], sequences[i], "reader %d did not share the first refresh", i)
	}
	assert.NotZero(t, sequences[0], "concurrent refresh must assign a sequence")
}

func TestFileScannerCacheSequence(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.log"), nil, 0o600), "failed to create initial file")

	c := newDirCache()
	newScanner := func() *fileScanner {
		return &fileScanner{
			dirCache:          c,
			dirCacheSequences: make(map[string]dirCacheSequence),
		}
	}
	first := newScanner()
	second := newScanner()

	names, err := first.readNames(dir, true)
	require.NoError(t, err, "first scanner failed")
	require.Equal(t, []string{"a.log"}, names, "initial listing differs")

	require.NoError(t, os.Rename(filepath.Join(dir, "a.log"), filepath.Join(dir, "b.log")), "failed to rename file")

	names, err = second.readNames(dir, true)
	require.NoError(t, err, "second scanner failed")
	assert.Equal(t, []string{"a.log"}, names, "a different scanner should share the current listing")

	names, err = first.readNames(dir, true)
	require.NoError(t, err, "first scanner failed to refresh")
	assert.Equal(t, []string{"b.log"}, names, "a scanner must not reuse its prior scan's listing")

	names, err = second.readNames(dir, true)
	require.NoError(t, err, "second scanner failed to consume refreshed listing")
	assert.Equal(t, []string{"b.log"}, names, "second scanner did not share the refreshed listing")
}

// ---- singleton lifecycle ----------------------------------------------------

func TestAcquireSharedDirReaderSingleton(t *testing.T) {
	// Require clean global state — skip if another test left a reference.
	dirReaderMu.Lock()
	if dirReaderInst != nil {
		dirReaderMu.Unlock()
		t.Skip("singleton already held by another test")
	}
	dirReaderMu.Unlock()

	dc1, release1 := acquireSharedDirReader()
	dc2, release2 := acquireSharedDirReader()

	dirReaderMu.Lock()
	assert.Equal(t, 2, dirReaderRefs, "two acquires must increment the reference count")
	assert.Same(t, dc1, dc2, "both acquires must return the same cache")
	dirReaderMu.Unlock()

	release1()
	dirReaderMu.Lock()
	assert.Equal(t, 1, dirReaderRefs, "first release must decrement the reference count")
	assert.NotNil(t, dirReaderInst, "cache must remain while a reference exists")
	dirReaderMu.Unlock()

	release2()
	dirReaderMu.Lock()
	assert.Equal(t, 0, dirReaderRefs, "last release must clear the reference count")
	assert.Nil(t, dirReaderInst, "cache must be nil after the last release")
	dirReaderMu.Unlock()
}

func TestAcquireSharedDirReaderReleaseIdempotent(t *testing.T) {
	dirReaderMu.Lock()
	if dirReaderInst != nil {
		dirReaderMu.Unlock()
		t.Skip("singleton already held by another test")
	}
	dirReaderMu.Unlock()

	_, release := acquireSharedDirReader()
	release()
	release()

	dirReaderMu.Lock()
	assert.Equal(t, 0, dirReaderRefs, "release must be idempotent")
	assert.Nil(t, dirReaderInst, "cache must be nil after release")
	dirReaderMu.Unlock()
}

func TestAcquireSharedDirReaderResetOnLastRelease(t *testing.T) {
	dirReaderMu.Lock()
	if dirReaderInst != nil {
		dirReaderMu.Unlock()
		t.Skip("singleton already held by another test")
	}
	dirReaderMu.Unlock()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.log"), nil, 0o600), "failed to create initial file")

	dc, release := acquireSharedDirReader()
	_, _, err := dc.readDirNames(dir, 0)
	require.NoError(t, err, "failed to warm directory cache")
	require.NotEmpty(t, dc.entries, "cache should contain the warmed directory")

	release()

	dirReaderMu.Lock()
	assert.Nil(t, dirReaderInst, "cache must be nil after the last release")
	dirReaderMu.Unlock()

	dc.mu.Lock()
	assert.Empty(t, dc.entries, "reset must clear all cached entries")
	dc.mu.Unlock()
}

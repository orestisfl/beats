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

//go:build linux

package filestream

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/elastic/elastic-agent-libs/logp"
)

// nonTmpfsTempDir returns a fresh directory under the current working
// directory, registered for cleanup with the test. The kernel page cache
// behaviour we exercise here does not apply to tmpfs (where the default
// os.TempDir lives on most Linux distros), so we cannot use t.TempDir().
func nonTmpfsTempDir(t testing.TB) string {
	t.Helper()
	dir, err := os.MkdirTemp(".", "fadvise-test-*")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return dir
}

// isFilesystemPageCached returns true if the given path lives on a
// filesystem whose pages participate in the OS page cache (i.e. not tmpfs).
// We use this to skip tests when the working directory happens to be on
// tmpfs.
func isFilesystemPageCached(t testing.TB, path string) bool {
	t.Helper()
	var st unix.Statfs_t
	require.NoError(t, unix.Statfs(path, &st))
	// 0x01021994 = TMPFS_MAGIC, 0x858458F6 = RAMFS_MAGIC
	const (
		tmpfsMagic = 0x01021994
		ramfsMagic = 0x858458F6
	)
	switch uint32(st.Type) { //nolint:gosec // magic numbers fit in uint32
	case tmpfsMagic, ramfsMagic:
		return false
	}
	return true
}

// cachedBytes returns the number of bytes of the file that are resident in
// the kernel page cache. It mmaps the file read-only and calls mincore(2)
// over the mapping. Note that this also touches the file via the mmap, but
// mincore reports pages that are in the page cache regardless of whether
// the calling process has them faulted in.
func cachedBytes(t testing.TB, path string) int64 {
	t.Helper()

	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	fi, err := f.Stat()
	require.NoError(t, err)
	size := fi.Size()
	if size == 0 {
		return 0
	}

	data, err := unix.Mmap(int(f.Fd()), 0, int(size), unix.PROT_READ, unix.MAP_SHARED) //nolint:gosec // fd and size fit
	require.NoError(t, err)
	defer func() { _ = unix.Munmap(data) }()

	pageSize := int64(os.Getpagesize())
	nPages := (size + pageSize - 1) / pageSize
	vec := make([]byte, nPages)
	_, _, errno := syscall.Syscall(
		unix.SYS_MINCORE,
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(size), //nolint:gosec // size is the bounded file size
		uintptr(unsafe.Pointer(&vec[0])),
	)
	require.Zero(t, errno, "mincore failed")

	var resident int64
	for i, b := range vec {
		if b&1 == 0 {
			continue
		}
		// Last page may be a partial page; account only the bytes that
		// belong to the file.
		pageEnd := int64(i+1) * pageSize
		if pageEnd > size {
			resident += size - int64(i)*pageSize
		} else {
			resident += pageSize
		}
	}
	return resident
}

// dropCacheForFile drops cached pages for the file by calling
// posix_fadvise(POSIX_FADV_DONTNEED) on the whole file. This avoids requiring
// CAP_SYS_ADMIN to write to /proc/sys/vm/drop_caches and gives us a clean
// baseline before each measurement.
func dropCacheForFile(t testing.TB, path string) {
	t.Helper()
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()
	require.NoError(t, fadviseDontNeed(f, 0, 0))
}

// syncFile flushes dirty pages of the file to disk so that
// posix_fadvise(POSIX_FADV_DONTNEED) can subsequently evict them. Without
// this, pages just written by os.WriteFile remain dirty and the kernel
// refuses to evict them.
func syncFile(t testing.TB, path string) {
	t.Helper()
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()
	require.NoError(t, f.Sync())
}

// TestFileCacheAdvise_FingerprintRange exercises the prospector fingerprint
// path in isolation: it runs a scanner with fingerprinting enabled and
// verifies that file_cache_advise causes the fingerprinted byte range to be
// evicted from the page cache after the scan, while leaving it cached when
// disabled.
func TestFileCacheAdvise_FingerprintRange(t *testing.T) {
	for _, advise := range []bool{false, true} {
		name := "fadvise_off"
		if advise {
			name = "fadvise_on"
		}
		t.Run(name, func(t *testing.T) {
			dir := nonTmpfsTempDir(t)
			if !isFilesystemPageCached(t, dir) {
				t.Skip("page cache test requires a non-tmpfs filesystem")
			}
			// 64 KiB is enough to span several pages.
			path := filepath.Join(dir, "fingerprint.log")
			require.NoError(t, os.WriteFile(path, make([]byte, 64*1024), 0o600))
			syncFile(t, path)
			dropCacheForFile(t, path)

			cfg := fileWatcherConfig{
				Scanner: fileScannerConfig{
					Fingerprint: fingerprintConfig{
						Enabled: true,
						Offset:  0,
						Length:  DefaultFingerprintSize,
					},
				},
			}
			fw, err := newFileWatcher(
				logp.NewNopLogger(),
				[]string{filepath.Join(dir, "*.log")},
				cfg,
				CompressionNone,
				false,
				advise,
				mustPathIdentifier(false),
				mustSourceIdentifier("cache-advise-fp-test"),
			)
			require.NoError(t, err)

			files := fw.GetFiles()
			require.Len(t, files, 1, "expected exactly one matched file")

			cached := cachedBytes(t, path)
			t.Logf("cached after fingerprint: %d bytes", cached)

			if advise {
				// With fadvise enabled we disable read-ahead for the fd and
				// advise the rounded-up fingerprint range as DONTNEED, so
				// no pages should remain cached.
				assert.Zero(t, cached,
					"with file_cache_advise the fingerprint range should be evicted")
			} else {
				// Without fadvise the kernel both reads the page and
				// speculatively reads-ahead more pages, all of which stay
				// cached.
				assert.GreaterOrEqual(t, cached, int64(os.Getpagesize()),
					"without file_cache_advise the fingerprint pages should remain cached")
			}
		})
	}
}

// TestFileCacheAdvise_EvictsPageCache runs the filestream input over a
// sizeable file with and without file_cache_advise, then measures how many
// of the file's pages are still resident in the OS page cache. With the
// option enabled, residency should drop to near zero; without it, the
// whole file should remain cached.
func TestFileCacheAdvise_EvictsPageCache(t *testing.T) {
	const lineCount = 200_000 // ~10 MiB at ~50B/line

	for _, advise := range []bool{false, true} {
		name := "fadvise_off"
		if advise {
			name = "fadvise_on"
		}
		t.Run(name, func(t *testing.T) {
			dir := nonTmpfsTempDir(t)
			if !isFilesystemPageCached(t, dir) {
				t.Skip("page cache test requires a non-tmpfs filesystem")
			}
			path := generateFile(t, dir, lineCount)

			// Make sure the written pages are clean (not dirty) so that
			// fadvise(POSIX_FADV_DONTNEED) can actually evict them.
			syncFile(t, path)
			dropCacheForFile(t, path)

			beforeCached := cachedBytes(t, path)
			t.Logf("cached before run: %d bytes", beforeCached)

			cfg := fmt.Sprintf(`
type: filestream
prospector.scanner.check_interval: 100ms
prospector.scanner.fingerprint.enabled: false
close.reader.on_eof: true
file_identity.native: ~
file_cache_advise: %t
paths:
  - %s
`, advise, filepath.Join(dir, "*"))

			runner := createFilestreamTestRunner(t, logp.NewNopLogger(),
				fmt.Sprintf("cache-advise-%s", name), cfg, int64(lineCount), false)
			runner(t)

			afterCached := cachedBytes(t, path)
			fi, err := os.Stat(path)
			require.NoError(t, err)
			fileSize := fi.Size()
			t.Logf("file size: %d bytes; cached after run: %d bytes (%.1f%%)",
				fileSize, afterCached, 100*float64(afterCached)/float64(fileSize))

			if advise {
				// With fadvise enabled we expect almost no pages to remain in
				// the page cache. Allow some slack for the active read-ahead
				// window that may still be present at the moment the harvester
				// is torn down.
				assert.Less(t, afterCached, fileSize/10,
					"with file_cache_advise enabled, less than 10%% of the file should remain cached")
			} else {
				// Without fadvise, the kernel keeps the whole file in cache
				// after sequential reads on a fresh page cache.
				assert.GreaterOrEqual(t, afterCached, fileSize/2,
					"without file_cache_advise, most of the file should remain cached")
			}
		})
	}
}

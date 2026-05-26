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
	"os"

	"golang.org/x/sys/unix"
)

// fadviseDontNeedSupported is set by build tag.
const fadviseDontNeedSupported = true

// fadviseDontNeed advises the kernel that the bytes in the range
// [offset, offset+length) of the file are no longer needed and the
// corresponding cached pages may be freed. A length of 0 means "from offset
// to end of file".
//
// On Linux this calls posix_fadvise(2) with POSIX_FADV_DONTNEED. The call
// is best-effort: the kernel only evicts clean, unmapped pages, and may
// ignore the advice entirely. On filesystems that do not implement
// posix_fadvise the syscall is a no-op and returns success.
//
// NOTE: the Linux kernel deliberately preserves *partial* pages at the start
// and end of the range (see mm/fadvise.c). Callers that want a sub-page
// range to be evicted must round length up to the page size, otherwise the
// call is a no-op for that range.
func fadviseDontNeed(f *os.File, offset, length int64) error {
	return unix.Fadvise(int(f.Fd()), offset, length, unix.FADV_DONTNEED) //nolint:gosec // fd from os.File always fits in int
}

// fadviseRandom hints the kernel that this fd will be accessed in random
// order, which disables read-ahead for the file struct. This is used before
// short fixed-range reads (e.g. prospector fingerprint reads) to prevent
// the kernel from populating the page cache with read-ahead pages that we
// would otherwise have to evict separately. The hint is per-fd and goes
// away when the fd is closed, so it has no impact on other readers of the
// same inode.
func fadviseRandom(f *os.File) error {
	return unix.Fadvise(int(f.Fd()), 0, 0, unix.FADV_RANDOM) //nolint:gosec // fd from os.File always fits in int
}

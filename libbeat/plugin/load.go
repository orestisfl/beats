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

//go:build (linux || darwin) && cgo

package plugin

import "errors"

// NOTE(DCE demo): importing the standard library "plugin" package forces the Go
// linker into degraded dead-code elimination. Once "plugin" is reachable the
// linker must retain every exported method of every reachable type, because a
// dynamically loaded .so could invoke any method through an interface at
// runtime. That single transitive import (pulled into every beat via
// libbeat/cmd/instance) defeats method-level DCE process-wide and inflates the
// binary. For this demo we drop the "plugin" import entirely; loadable plugin
// support is disabled so the linker can perform method DCE again.
func loadPlugins(path string) error {
	return errors.New("loadable plugins disabled in this build")
}

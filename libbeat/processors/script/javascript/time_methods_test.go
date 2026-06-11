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
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
// either express or implied. See the License for the
// specific language governing permissions and limitations
// under the License.

package javascript

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/elastic-agent-libs/mapstr"
)

// TestTimeMethods verifies that the time.Time returned by event.Get("@timestamp")
// exposes the accessors scripts rely on (registered via goja.RegisterNativeMethods),
// even under the default goja build where reflective method exposure is disabled.
func TestTimeMethods(t *testing.T) {
	ts := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	p := newTestProcessor(t, "source", `
		function process(evt) {
			var ts = evt.Get("@timestamp");
			evt.Put("enriched.epoch", ts.Unix());
			evt.Put("enriched.year", ts.Year());
			evt.Put("enriched.month", ts.Month());
		}`, "")

	result, err := p.Run(&beat.Event{Timestamp: ts, Fields: mapstr.M{}})
	require.NoError(t, err)

	epoch, _ := result.GetValue("enriched.epoch")
	assert.Equal(t, int64(ts.Unix()), epoch)
	year, _ := result.GetValue("enriched.year")
	assert.Equal(t, int64(2009), year)
	month, _ := result.GetValue("enriched.month")
	assert.Equal(t, int64(11), month)
}

// TestTimeRoundTrip verifies the wrapped @timestamp is still a genuine time.Time,
// so reading it and writing it back preserves the event Timestamp (Export() must
// yield time.Time, not a wrapper, to satisfy setTimestamp).
func TestTimeRoundTrip(t *testing.T) {
	ts := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	p := newTestProcessor(t, "source", `
		function process(evt) {
			evt.Put("@timestamp", evt.Get("@timestamp"));
		}`, "")

	result, err := p.Run(&beat.Event{Timestamp: ts, Fields: mapstr.M{}})
	require.NoError(t, err)
	assert.True(t, result.Timestamp.Equal(ts), "timestamp changed: got %v want %v", result.Timestamp, ts)
}

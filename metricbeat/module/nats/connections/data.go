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

package connections

import (
	"encoding/json"
	"fmt"

	s "github.com/elastic/beats/v8/libbeat/common/schema"
	c "github.com/elastic/beats/v8/libbeat/common/schema/mapstriface"
	"github.com/elastic/beats/v8/metricbeat/mb"
	"github.com/elastic/beats/v8/metricbeat/module/nats/util"
)

var (
	moduleSchema = s.Schema{
		"server": s.Object{
			"id":   c.Str("server_id"),
			"time": c.Str("now"),
		},
	}
	connectionsSchema = s.Schema{
		"total": c.Int("total"),
	}
)

func eventMapping(r mb.ReporterV2, content []byte) error {
	var inInterface map[string]interface{}

	err := json.Unmarshal(content, &inInterface)
	if err != nil {
		return fmt.Errorf("failure parsing NATS connections API response: %w", err)
	}
	metricSetFields, err := connectionsSchema.Apply(inInterface)
	if err != nil {
		return fmt.Errorf("failure applying connections schema: %w", err)

	}

	moduleFields, err := moduleSchema.Apply(inInterface)
	if err != nil {
		return fmt.Errorf("failure applying module schema: %w", err)
	}
	timestamp, err := util.GetNatsTimestamp(moduleFields)
	if err != nil {
		return fmt.Errorf("failure parsing server timestamp: %w", err)
	}
	event := mb.Event{
		MetricSetFields: metricSetFields,
		ModuleFields:    moduleFields,
		Timestamp:       timestamp,
	}
	r.Event(event)
	return nil
}

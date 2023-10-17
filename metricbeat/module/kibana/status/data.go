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

package status

import (
	"encoding/json"
	"fmt"

	s "github.com/elastic/beats/v8/libbeat/common/schema"
	c "github.com/elastic/beats/v8/libbeat/common/schema/mapstriface"
	"github.com/elastic/beats/v8/metricbeat/helper/elastic"
	"github.com/elastic/beats/v8/metricbeat/mb"
	"github.com/elastic/beats/v8/metricbeat/module/kibana"
	"github.com/elastic/elastic-agent-libs/mapstr"
)

var (
	schema = s.Schema{
		"uuid": c.Str("uuid"),
		"name": c.Str("name"),
		"version": c.Dict("version", s.Schema{
			"number": c.Str("number"),
		}),
		"status": c.Dict("status", s.Schema{
			"overall": c.Dict("overall", s.Schema{
				"state": c.Str("state"),
			}),
		}),
		"metrics": c.Dict("metrics", s.Schema{
			"requests": c.Dict("requests", s.Schema{
				"total":       c.Int("total"),
				"disconnects": c.Int("disconnects"),
			}),
			"concurrent_connections": c.Int("concurrent_connections"),
		}),
	}
)

func eventMapping(r mb.ReporterV2, content []byte, isXpack bool) error {
	var event mb.Event
	event.RootFields = mapstr.M{}
	event.RootFields.Put("service.name", kibana.ModuleName)

	var data map[string]interface{}
	err := json.Unmarshal(content, &data)
	if err != nil {
		return fmt.Errorf("failure parsing Kibana Status API response: %w", err)
	}

	dataFields, _ := schema.Apply(data)

	// Set service ID
	uuid, err := dataFields.GetValue("uuid")
	if err != nil {
		return elastic.MakeErrorForMissingField("uuid", elastic.Kibana)
	}
	event.RootFields.Put("service.id", uuid)
	dataFields.Delete("uuid")

	// Set service version
	version, err := dataFields.GetValue("version.number")
	if err != nil {
		return elastic.MakeErrorForMissingField("version.number", elastic.Kibana)
	}
	event.RootFields.Put("service.version", version)
	dataFields.Delete("version")

	event.MetricSetFields = dataFields

	// xpack.enabled in config using standalone metricbeat writes to `.monitoring` instead of `metricbeat-*`
	// When using Agent, the index name is overwritten anyways.
	if isXpack {
		index := elastic.MakeXPackMonitoringIndexName(elastic.Kibana)
		event.Index = index
	}

	r.Event(event)
	return nil
}

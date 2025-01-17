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

package settings

import (
	"encoding/json"
	"fmt"

	"github.com/elastic/beats/v8/metricbeat/helper/elastic"
	"github.com/elastic/elastic-agent-libs/mapstr"

	s "github.com/elastic/beats/v8/libbeat/common/schema"
	c "github.com/elastic/beats/v8/libbeat/common/schema/mapstriface"
	"github.com/elastic/beats/v8/metricbeat/mb"
)

func eventMapping(r mb.ReporterV2, content []byte) error {
	var data map[string]interface{}
	err := json.Unmarshal(content, &data)
	if err != nil {
		return fmt.Errorf("failure parsing Kibana API response: %w", err)
	}

	schema := s.Schema{
		"elasticsearch": s.Object{
			"cluster": s.Object{
				"id": c.Str("cluster_uuid"),
			},
		},
		"settings": c.Ifc("settings.kibana"),
	}

	res, err := schema.Apply(data)
	if err != nil {
		return err
	}

	event := mb.Event{
		ModuleFields:    res,
		MetricSetFields: nil,
		RootFields:      make(mapstr.M),
	}

	// Set service address
	serviceAddress, err := res.GetValue("settings.transport_address")
	if err != nil {
		event.Error = elastic.MakeErrorForMissingField("kibana.transport_address", elastic.Kibana)
		return event.Error
	}
	event.RootFields.Put("service.address", serviceAddress)

	r.Event(event)

	return nil
}

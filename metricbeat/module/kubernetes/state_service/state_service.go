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

package state_service

import (
	"github.com/elastic/beats/v8/metricbeat/helper/kubernetes"
	p "github.com/elastic/beats/v8/metricbeat/helper/prometheus"
	"github.com/elastic/beats/v8/metricbeat/mb"
	"github.com/elastic/beats/v8/metricbeat/module/kubernetes/util"
)

// mapping stores the state metrics we want to fetch and will be used by this metricset
var mapping = &p.MetricsMapping{
	Metrics: map[string]p.MetricMap{
		"kube_service_info": p.InfoMetric(),
		"kube_service_labels": p.ExtendedInfoMetric(
			p.Configuration{
				StoreNonMappedLabels:     true,
				NonMappedLabelsPlacement: mb.ModuleDataKey + ".labels",
				MetricProcessingOptions:  []p.MetricOption{p.OpLabelKeyPrefixRemover("label_")},
			}),
		"kube_service_created":                      p.Metric("created", p.OpUnixTimestampValue()),
		"kube_service_spec_type":                    p.InfoMetric(),
		"kube_service_spec_external_ip":             p.InfoMetric(),
		"kube_service_status_load_balancer_ingress": p.InfoMetric(),
	},
	Labels: map[string]p.LabelMap{
		"namespace":        p.KeyLabel(mb.ModuleDataKey + ".namespace"),
		"service":          p.KeyLabel("name"),
		"cluster_ip":       p.Label("cluster_ip"),
		"external_name":    p.Label("external_name"),
		"external_ip":      p.Label("external_ip"),
		"load_balancer_ip": p.Label("load_balancer_ip"),
		"type":             p.Label("type"),
		"ip":               p.Label("ingress_ip"),
		"hostname":         p.Label("ingress_hostname"),
	},
}

// Register metricset
func init() {
	kubernetes.Init(util.ServiceResource, mapping)
}

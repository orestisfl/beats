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

// Code generated by beats/dev-tools/cmd/asset/asset.go - DO NOT EDIT.

package nginx

import (
	"github.com/elastic/beats/v8/libbeat/asset"
)

func init() {
	if err := asset.SetFields("metricbeat", "nginx", asset.ModuleFieldsPri, AssetNginx); err != nil {
		panic(err)
	}
}

// AssetNginx returns asset data.
// This is the base64 encoded zlib format compressed contents of module/nginx.
func AssetNginx() string {
	return "eJzElM9u2zAMxu9+ig8+r3kAH/YIOw3YYRhSVaJtIYrkkXTSvP0gR26cwC1SoMF0skWJ349/xCfs6NQgdj6+VoB6DdSg/pH/6wpwJJb9oD7FBt8rAJhsEOIDMUSNjoI9KXsrsCkEskoOLac9DoZ9yubkxkCyqQDpE+vWptj6rkFrglAFMAUyQg06k8+Qqo+dNPhdi4T6G+pedaj/VEDrKThpJpInRLOnC31eehqyG07jUHZWQsjrebr1DJuiGh8F2tNbHNobxZGYIJbNMMczXdkUF0uSJY3o+HJOy5tpDesDtAnv4uZzjPlI7F63OWPb7GR79rI9F2Gu2GA62iwUryswr9sYl3H2STR/XRnnSHd0OiZ2N7YP4r101ux3s6pqrPrDumZIsfuc4M+eYEdmioo47l+IkdoiARt83rcpRrLZg8BHG0bnY4dfxucWXVrf47U0qHwhsCY14Qo3K5BbAV5H6k10gW5r85VIReFuIsdpGB5KVBTuJmL6O5I8tG4FZVZa5yjd+dB+vwuEyeTGfyzI4qkd+zzZzhPByyw/TbcCip6MI17HPfL0Pv8XbpEvuDKkKIQXY3fQNG2ec/4Ou3k8u3dhdcIVbbSJYeZMb6p/AQAA//9pr1BQ"
}

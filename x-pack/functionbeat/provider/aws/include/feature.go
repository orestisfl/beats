// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package include

import (
	"github.com/elastic/beats/v8/libbeat/feature"
	"github.com/elastic/beats/v8/x-pack/functionbeat/function/provider"
	"github.com/elastic/beats/v8/x-pack/functionbeat/provider/aws/aws"
)

// Bundle exposes the trigger supported by the AWS provider.
var features = provider.Builder(
	"aws",
	provider.NewDefaultProvider("aws", provider.NewNullCli, provider.NewNullTemplateBuilder),
	feature.MakeDetails("AWS Lambda", "listen to events on AWS lambda", feature.Stable),
).AddFunction("cloudwatch_logs",
	aws.NewCloudwatchLogs,
	aws.CloudwatchLogsDetails(),
).AddFunction("api_gateway_proxy",
	aws.NewAPIGatewayProxy,
	aws.APIGatewayProxyDetails(),
).AddFunction("kinesis",
	aws.NewKinesis,
	aws.KinesisDetails(),
).AddFunction("sqs",
	aws.NewSQS,
	aws.SQSDetails(),
).AddFunction("cloudwatch_logs_kinesis",
	aws.NewCloudwatchKinesis,
	aws.CloudwatchKinesisDetails(),
).Features()

func init() {
	feature.MustRegister(features...)
}

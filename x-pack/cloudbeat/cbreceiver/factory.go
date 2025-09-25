// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package cbreceiver

import (
	"context"
	"fmt"

	"github.com/elastic/beats/v7/libbeat/cmd/instance"
	"github.com/elastic/beats/v7/libbeat/monitoring/report"
	xpInstance "github.com/elastic/beats/v7/x-pack/libbeat/cmd/instance"
	"github.com/spf13/pflag"

	cloudbeat "github.com/elastic/cloudbeat/pkg/beater"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

const (
	Name = "cloudbeatreceiver"
)

func createDefaultConfig() component.Config {
	return &Config{}
}

func createReceiver(_ context.Context, set receiver.Settings, baseCfg component.Config, consumer consumer.Logs) (receiver.Logs, error) {
	cfg, ok := baseCfg.(*Config)
	if !ok {
		return nil, fmt.Errorf("could not convert otel config to cloudbeat config")
	}

	b, err := xpInstance.NewBeatForReceiver(
		cloudbeatSettings(),
		cfg.Beatconfig,
		true,
		consumer,
		set.ID.String(),
		set.Logger.Core(),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating %s: %w", Name, err)
	}

	br, err := xpInstance.NewBeatReceiver(b, cloudbeat.NewBeater)
	if err != nil {
		return nil, fmt.Errorf("error creating %s:%w", Name, err)
	}

	return &cloudbeatReceiver{BeatReceiver: br}, nil
}

func cloudbeatSettings() instance.Settings {
	runFlags := pflag.NewFlagSet(Name, pflag.ExitOnError)
	return instance.Settings{
		Name:            Name,
		HasDashboards:   false,
		ElasticLicensed: true,
		Monitoring:      report.Settings{},
		RunFlags:        runFlags,
	}
}

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType(Name),
		createDefaultConfig,
		receiver.WithLogs(createReceiver, component.StabilityLevelAlpha))
}

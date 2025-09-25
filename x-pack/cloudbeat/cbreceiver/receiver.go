// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package cbreceiver

import (
	"context"
	"fmt"
	"sync"

	xpInstance "github.com/elastic/beats/v7/x-pack/libbeat/cmd/instance"

	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

type cloudbeatReceiver struct {
	xpInstance.BeatReceiver
	wg sync.WaitGroup
}

func (cb *cloudbeatReceiver) Start(_ context.Context, host component.Host) error {
	cb.wg.Add(1)
	go func() {
		defer cb.wg.Done()
		cb.Logger.Info("starting cloudbeat receiver")
		if err := cb.BeatReceiver.Start(host); err != nil {
			cb.Logger.Error("error starting cloudbeat receiver", zap.Error(err))
		}
	}()
	return nil
}

func (cb *cloudbeatReceiver) Shutdown(context.Context) error {
	cb.Logger.Info("stopping cloudbeat receiver")
	if err := cb.BeatReceiver.Shutdown(); err != nil {
		return fmt.Errorf("error stopping cloudbeat receiver: %w", err)
	}
	cb.wg.Wait()
	return nil
}

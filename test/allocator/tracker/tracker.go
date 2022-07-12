/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package tracker

import (
	"time"

	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter"
	"github.com/newcloudtechnologies/memlimiter/utils/breaker"
	"github.com/pkg/errors"
)

// Tracker is responsible for service stats persistence.
type Tracker struct {
	backend    backend
	memLimiter memlimiter.Service
	cfg        *Config
	breaker    *breaker.Breaker
	logger     logr.Logger
}

func (tr *Tracker) makeReport() (*Report, error) {
	out := &Report{}

	out.Timestamp = time.Now().Format(time.RFC3339Nano)

	mlStats, err := tr.memLimiter.GetStats()
	if err != nil {
		return nil, errors.Wrap(err, "memlimiter stats")
	}

	if mlStats != nil {
		out.RSS = mlStats.Controller.MemoryBudget.RSSActual
		out.Utilization = mlStats.Controller.MemoryBudget.Utilization

		if mlStats.Backpressure != nil {
			out.GOGC = mlStats.Backpressure.ControlParameters.GOGC
			out.Throttling = mlStats.Backpressure.ControlParameters.ThrottlingPercentage
		}
	}

	return out, nil
}

func (tr *Tracker) dumpReport() error {
	r, err := tr.makeReport()
	if err != nil {
		return errors.Wrap(err, "dump Report")
	}

	if err = tr.backend.saveReport(r); err != nil {
		return errors.Wrap(err, "backend save Report")
	}

	return nil
}

func (tr *Tracker) loop() {
	defer tr.breaker.Dec()

	ticker := time.NewTicker(tr.cfg.Period.Duration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := tr.dumpReport(); err != nil {
				tr.logger.Error(err, "dump Report")
			}
		case <-tr.breaker.Done():
			return
		}
	}
}

func (tr *Tracker) GetReports() ([]*Report, error) { return tr.backend.getReports() }

// Quit gracefully terminates tracker.
func (tr *Tracker) Quit() {
	tr.breaker.ShutdownAndWait()
	tr.backend.quit()
}

// NewTrackerFromConfig is a constructor of a Tracker.
func NewTrackerFromConfig(logger logr.Logger, cfg *Config, memLimiter memlimiter.Service) (*Tracker, error) {
	var (
		back backend
		err  error
	)
	switch {
	case cfg.BackendFile != nil:
		back, err = newBackendFile(logger, cfg.BackendFile)
	case cfg.BackendMemory != nil:
		back = newBackendMemory()
	default:
		return nil, errors.New("unexpected backend type")
	}

	if err != nil {
		return nil, errors.Wrap(err, "new backend")
	}

	tr := &Tracker{
		backend:    back,
		logger:     logger,
		cfg:        cfg,
		memLimiter: memLimiter,
		breaker:    breaker.NewBreakerWithInitValue(1),
	}

	go tr.loop()

	return tr, nil
}

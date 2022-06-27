/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package tracker

import (
	"encoding/csv"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter"
	"github.com/newcloudtechnologies/memlimiter/utils/breaker"
	"github.com/pkg/errors"
)

// Tracker is responsible for service stats persistence.
type Tracker struct {
	memLimiter memlimiter.Service
	writer     *csv.Writer
	cfg        *Config
	fd         *os.File
	breaker    *breaker.Breaker
	logger     logr.Logger
}

func (tr *Tracker) makeReport() (*report, error) {
	out := &report{}

	out.timestamp = time.Now().Format(time.RFC3339Nano)

	mlStats, err := tr.memLimiter.GetStats()
	if err != nil {
		return nil, errors.Wrap(err, "memlimiter stats")
	}

	if mlStats != nil {
		out.rss = mlStats.Controller.MemoryBudget.RSSActual
		out.utilization = mlStats.Controller.MemoryBudget.Utilization

		if mlStats.Backpressure != nil {
			out.gogc = mlStats.Backpressure.ControlParameters.GOGC
			out.throttling = mlStats.Backpressure.ControlParameters.ThrottlingPercentage
		}
	}

	return out, nil
}

func (tr *Tracker) dumpReport() error {
	r, err := tr.makeReport()
	if err != nil {
		return errors.Wrap(err, "dump report")
	}

	if err := tr.writer.Write(r.toCsv()); err != nil {
		return errors.Wrap(err, "csv write")
	}

	tr.writer.Flush()

	if err := tr.writer.Error(); err != nil {
		return errors.Wrap(err, "csv flush")
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
				tr.logger.Error(err, "dump report")
			}
		case <-tr.breaker.Done():
			return
		}
	}
}

// Quit gracefully terminates tracker.
func (tr *Tracker) Quit() {
	tr.breaker.ShutdownAndWait()

	if err := tr.fd.Close(); err != nil {
		tr.logger.Error(err, "close file")
	}
}

// NewTrackerFromConfig is a constructor of a Tracker.
func NewTrackerFromConfig(logger logr.Logger, cfg *Config, memLimiter memlimiter.Service) (*Tracker, error) {
	const perm = 0600

	fd, err := os.OpenFile(cfg.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY|os.O_SYNC|os.O_TRUNC, perm)
	if err != nil {
		return nil, errors.Wrap(err, "open file")
	}

	wr := csv.NewWriter(fd)

	if err := wr.Write(new(report).headers()); err != nil {
		return nil, errors.Wrap(err, "write header")
	}

	tr := &Tracker{
		logger:     logger,
		fd:         fd,
		cfg:        cfg,
		writer:     wr,
		memLimiter: memLimiter,
		breaker:    breaker.NewBreakerWithInitValue(1),
	}

	go tr.loop()

	return tr, nil
}

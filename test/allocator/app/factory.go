/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package app

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter/test/allocator/perf"
	"github.com/newcloudtechnologies/memlimiter/test/allocator/server"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

// Runnable represents some task that can be run.
type Runnable interface {
	// Run - a blocking call.
	Run() error
	// Quit terminates process.
	Quit()
}

// Factory builds runnable tasks.
type Factory interface {
	// MakeServer creates a server.
	MakeServer(c *cli.Context) (Runnable, error)
	// MakePerfClient creates a client for performance tests.
	MakePerfClient(c *cli.Context) (Runnable, error)
}

type factoryDefault struct {
	logger logr.Logger
}

//nolint:dupl
func (f *factoryDefault) MakeServer(c *cli.Context) (Runnable, error) {
	filename := c.String("config")

	data, err := ioutil.ReadFile(filepath.Clean(filename))
	if err != nil {
		return nil, errors.Wrap(err, "ioutil readfile")
	}

	cfg := &server.Config{}

	if err = json.Unmarshal(data, cfg); err != nil {
		return nil, errors.Wrap(err, "unmarshal")
	}

	srv, err := server.NewAllocatorServer(f.logger, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "new allocator server")
	}

	return srv, nil
}

//nolint:dupl
func (f *factoryDefault) MakePerfClient(c *cli.Context) (Runnable, error) {
	filename := c.String("config")

	data, err := ioutil.ReadFile(filepath.Clean(filename))
	if err != nil {
		return nil, errors.Wrap(err, "ioutil readfile")
	}

	cfg := &perf.Config{}

	if err = json.Unmarshal(data, cfg); err != nil {
		return nil, errors.Wrap(err, "unmarshal")
	}

	cl, err := perf.NewClient(f.logger, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "new allocator server")
	}

	return cl, nil
}

// NewFactory makes new default factory.
func NewFactory(logger logr.Logger) Factory {
	return &factoryDefault{logger: logger}
}

/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/newcloudtechnologies/memlimiter/test/allocator/perf"
	"github.com/newcloudtechnologies/memlimiter/test/allocator/server"
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

	data, err := os.ReadFile(filepath.Clean(filename))
	if err != nil {
		return nil, fmt.Errorf("os readfile: %w", err)
	}

	cfg := &server.Config{}

	if err = json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	srv, err := server.NewServer(f.logger, cfg)
	if err != nil {
		return nil, fmt.Errorf("new allocator server: %w", err)
	}

	return srv, nil
}

//nolint:dupl
func (f *factoryDefault) MakePerfClient(c *cli.Context) (Runnable, error) {
	filename := c.String("config")

	data, err := os.ReadFile(filepath.Clean(filename))
	if err != nil {
		return nil, fmt.Errorf("os readfile: %w", err)
	}

	cfg := &perf.Config{}

	if err = json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	cl, err := perf.NewClient(f.logger, cfg)
	if err != nil {
		return nil, fmt.Errorf("new allocator server: %w", err)
	}

	return cl, nil
}

// NewFactory makes new default factory.
func NewFactory(logger logr.Logger) Factory {
	return &factoryDefault{logger: logger}
}

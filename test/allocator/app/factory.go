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

type Runnable interface {
	Run() error
	Quit()
}

type Factory interface {
	MakeServer(c *cli.Context) (Runnable, error)
	MakePerfClient(c *cli.Context) (Runnable, error)
}

type factoryDefault struct {
	logger logr.Logger
}

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

func NewFactory(logger logr.Logger) Factory {
	return &factoryDefault{logger: logger}
}

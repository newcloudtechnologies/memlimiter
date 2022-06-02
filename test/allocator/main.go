package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/newcloudtechnologies/memlimiter/test/allocator/perf"
	"github.com/newcloudtechnologies/memlimiter/test/allocator/server"
)

func main() {
	logger := stdr.NewWithOptions(
		log.New(os.Stdout, "", log.LstdFlags),
		stdr.Options{LogCaller: stdr.All},
	)

	app := &cli.App{
		Name:  "allocator",
		Usage: "test application for memlimiter",
		Commands: cli.Commands{
			&cli.Command{
				Name:  "server",
				Usage: "allocator server app",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "config",
						Usage:    "configuration file",
						Aliases:  []string{"c"},
						Required: true,
					},
				},
				Action: func(context *cli.Context) error { return actionServer(logger, context) },
			},
			&cli.Command{
				Name:  "perf",
				Usage: "allocator perf client",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "config",
						Usage:    "configuration file",
						Aliases:  []string{"c"},
						Required: true,
					},
				},
				Action: actionPerf,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Error(err, "application run")
		os.Exit(1)
	}
}

func actionServer(logger logr.Logger, c *cli.Context) error {
	srv, err := makeServer(logger, c)
	if err != nil {
		return errors.Wrap(err, "make server")
	}

	if err := runAndWaitSignal(srv); err != nil {
		return errors.Wrap(err, "run and wait signal")
	}

	return nil
}

func makeServer(logger logr.Logger, c *cli.Context) (server.Server, error) {
	filename := c.String("config")

	data, err := ioutil.ReadFile(filepath.Clean(filename))
	if err != nil {
		return nil, errors.Wrap(err, "ioutil readfile")
	}

	cfg := &server.Config{}

	if err = json.Unmarshal(data, cfg); err != nil {
		return nil, errors.Wrap(err, "unmarshal")
	}

	srv, err := server.NewAllocatorServer(logger, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "new allocator server")
	}

	return srv, nil
}

func actionPerf(c *cli.Context) error {
	perfClient, err := makePerf(c)
	if err != nil {
		return errors.Wrap(err, "make perf")
	}

	if err := runAndWaitSignal(perfClient); err != nil {
		return errors.Wrap(err, "run and wait signal")
	}

	return nil
}

func makePerf(c *cli.Context) (*perf.Client, error) {
	filename := c.String("config")

	data, err := ioutil.ReadFile(filepath.Clean(filename))
	if err != nil {
		return nil, errors.Wrap(err, "ioutil readfile")
	}

	cfg := &perf.Config{}

	if err = json.Unmarshal(data, cfg); err != nil {
		return nil, errors.Wrap(err, "unmarshal")
	}

	srv, err := perf.NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "new allocator server")
	}

	return srv, nil
}

type runnable interface {
	Run() error
	Quit()
}

func runAndWaitSignal(r runnable) error {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	errChan := make(chan error, 1)

	go func() { errChan <- r.Run() }()

	defer r.Quit()

	select {
	case err := <-errChan:
		return errors.Wrap(err, "run error")
	case <-signalChan:
		return nil
	}
}

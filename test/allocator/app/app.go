package app

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

// App - CLI application.
type App struct {
	logger  logr.Logger
	factory Factory
}

// Run launches the application.
func (a *App) Run() {
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
				Action: func(context *cli.Context) error {
					r, err := a.factory.MakeServer(context)
					if err != nil {
						return errors.Wrap(err, "make server")
					}

					return a.runAndWaitSignal(r)
				},
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
				Action: func(context *cli.Context) error {
					r, err := a.factory.MakePerfClient(context)
					if err != nil {
						return errors.Wrap(err, "make perf client")
					}

					return a.runAndWaitSignal(r)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		a.logger.Error(err, "application run")
		os.Exit(1)
	}
}

func (a *App) runAndWaitSignal(r Runnable) error {
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

// NewApp prepares new application.
func NewApp(logger logr.Logger, factory Factory) *App {
	return &App{
		logger:  logger,
		factory: factory,
	}
}

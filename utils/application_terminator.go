/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package utils

import (
	"os"

	"github.com/go-logr/logr"
)

// ApplicationTerminator shuts down the application by MemLimiter command.
// It is used in the cases when it's better to restart the application rather than continue working.
// It must be implemented by library users, because every application has its own
// graceful termination protocol.
type ApplicationTerminator interface {
	// Terminate is a special method registering the fatal error of memory management.
	// It's mandatory for the application to terminate itself within or after this call.
	Terminate(fatalErr error)
}

type ungracefulApplicationTerminator struct {
	logger logr.Logger
}

func (at *ungracefulApplicationTerminator) Terminate(fatalErr error) {
	at.logger.Error(fatalErr, "terminate application due to fatal error")
	os.Exit(1)
}

// NewUngracefulApplicationTerminator creates trivial implementation of the ApplicationTerminator,
// which immediately calls os.Exit(1). Can be used only with simple and stateless services.
func NewUngracefulApplicationTerminator(logger logr.Logger) ApplicationTerminator {
	return &ungracefulApplicationTerminator{
		logger: logger,
	}
}

type fatalErrChanApplicationTerminator struct {
	fatalErrChan chan<- error
}

func (at *fatalErrChanApplicationTerminator) Terminate(fatalErr error) { at.fatalErrChan <- fatalErr }

// NewFatalErrChanApplicationTerminator creates an instance of the ApplicationTerminator
// that put fatal errors to special channels, so the channel reader can handle it in another thread.
func NewFatalErrChanApplicationTerminator(fatalErrChan chan<- error) ApplicationTerminator {
	return &fatalErrChanApplicationTerminator{
		fatalErrChan: fatalErrChan,
	}
}

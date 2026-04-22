/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package tracker

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"

	"github.com/go-logr/logr"
)

var _ backend = (*backendFile)(nil)

type backendFile struct {
	fd     *os.File
	writer *csv.Writer
	logger logr.Logger
}

func (b *backendFile) saveReport(r *Report) error {
	err := b.writer.Write(r.toCsv())
	if err != nil {
		return fmt.Errorf("csv write: %w", err)
	}

	b.writer.Flush()

	err = b.writer.Error()
	if err != nil {
		return fmt.Errorf("csv flush: %w", err)
	}

	return nil
}

func (b *backendFile) getReports() ([]*Report, error) {
	return nil, errors.New("all reports are dumped to file immediately")
}

func (b *backendFile) quit() {
	err := b.fd.Close()
	if err != nil {
		b.logger.Error(err, "close file")
	}
}

func newBackendFile(logger logr.Logger, cfg *ConfigBackendFile) (backend, error) {
	const perm = 0600

	fd, err := os.OpenFile(cfg.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY|os.O_SYNC|os.O_TRUNC, perm)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	wr := csv.NewWriter(fd)

	if err := wr.Write(new(Report).headers()); err != nil {
		return nil, fmt.Errorf("write header: %w", err)
	}

	return &backendFile{logger: logger, writer: wr, fd: fd}, nil
}

/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package tracker

import (
	"encoding/csv"
	"os"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

var _ backend = (*backendFile)(nil)

type backendFile struct {
	logger logr.Logger
	writer *csv.Writer
	fd     *os.File
}

func (b *backendFile) saveReport(r *Report) error {
	if err := b.writer.Write(r.toCsv()); err != nil {
		return errors.Wrap(err, "csv write")
	}

	b.writer.Flush()

	if err := b.writer.Error(); err != nil {
		return errors.Wrap(err, "csv flush")
	}

	return nil
}

func (b *backendFile) getReports() ([]*Report, error) {
	return nil, errors.New("all reports are dumped to file immediately")
}

func (b *backendFile) quit() {
	if err := b.fd.Close(); err != nil {
		b.logger.Error(err, "close file")
	}
}

func newBackendFile(logger logr.Logger, cfg *ConfigBackendFile) (backend, error) {
	const perm = 0600

	fd, err := os.OpenFile(cfg.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY|os.O_SYNC|os.O_TRUNC, perm)
	if err != nil {
		return nil, errors.Wrap(err, "open file")
	}

	wr := csv.NewWriter(fd)

	if err := wr.Write(new(Report).headers()); err != nil {
		return nil, errors.Wrap(err, "write header")
	}

	return &backendFile{logger: logger, writer: wr, fd: fd}, nil
}

/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package tracker

import (
	"testing"
	"time"

	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/require"
)

func TestBackend(t *testing.T) {
	logger := testr.New(t)

	cfg := &ConfigBackendFile{Path: "/tmp/backend.csv"}

	back, err := newBackendFile(logger, cfg)
	require.NoError(t, err)

	defer back.quit()

	reportsIn := []*Report{
		{
			Timestamp:   time.Now().String(),
			RSS:         1,
			Utilization: 2,
			GOGC:        3,
			Throttling:  4,
		},
		{
			Timestamp:   time.Now().String(),
			RSS:         2,
			Utilization: 3,
			GOGC:        4,
			Throttling:  5,
		},
	}

	for _, rep := range reportsIn {
		err = back.saveReport(rep)
		require.NoError(t, err)
	}

	reportsOut, err := back.getReports()
	require.Len(t, reportsOut, 0)
	require.Error(t, err)
}

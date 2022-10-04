/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package integration

import (
	"testing"
	"time"

	"code.cloudfoundry.org/bytefmt"
	"github.com/aclements/go-moremath/stats"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"github.com/newcloudtechnologies/memlimiter"
	"github.com/newcloudtechnologies/memlimiter/controller/nextgc"
	"github.com/newcloudtechnologies/memlimiter/test/allocator/perf"
	"github.com/newcloudtechnologies/memlimiter/test/allocator/server"
	"github.com/newcloudtechnologies/memlimiter/test/allocator/tracker"
	"github.com/newcloudtechnologies/memlimiter/utils/config/bytes"
	"github.com/newcloudtechnologies/memlimiter/utils/config/duration"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestComponent(t *testing.T) {
	const endpoint = "0.0.0.0:1988"

	logger := testr.New(t)

	const rssLimit = bytefmt.GIGABYTE

	allocatorServer, err := makeServer(logger, endpoint, rssLimit)
	require.NoError(t, err)

	defer allocatorServer.Quit()

	go func() {
		if errRun := allocatorServer.Run(); errRun != nil {
			logger.Error(errRun, "server run")
		}
	}()

	// wait for a while to make server run asynchronously
	time.Sleep(time.Second)

	perfClient, err := makePerfClient(logger, endpoint)
	require.NoError(t, err)

	// perform load
	err = perfClient.Run()
	require.NoError(t, err)

	defer perfClient.Quit()

	// collect reports
	reports, err := allocatorServer.Tracker().GetReports()
	require.NoError(t, err)
	require.NotEmpty(t, reports)

	analyzeReports(t, reports, rssLimit)
}

func makeServer(logger logr.Logger, endpoint string, rssLimit uint64) (server.Server, error) {
	cfg := &server.Config{
		MemLimiter: &memlimiter.Config{ControllerNextGC: &nextgc.ControllerConfig{
			RSSLimit:             bytes.Bytes{Value: rssLimit},
			DangerZoneGOGC:       50,
			DangerZoneThrottling: 90,
			Period:               duration.Duration{Duration: time.Second},
			ComponentProportional: &nextgc.ComponentProportionalConfig{
				Coefficient: 20,
				WindowSize:  20,
			},
		}},
		Tracker: &tracker.Config{
			BackendMemory: &tracker.ConfigBackendMemory{},
			Period:        duration.Duration{Duration: time.Second},
		},
		ListenEndpoint: endpoint,
	}

	allocatorServer, err := server.NewServer(logger, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "perf client")
	}

	return allocatorServer, nil
}

func makePerfClient(logger logr.Logger, endpoint string) (*perf.Client, error) {
	cfg := &perf.Config{
		Endpoint:       endpoint,
		RPS:            100,
		LoadDuration:   duration.Duration{Duration: 20 * time.Second},
		AllocationSize: bytes.Bytes{Value: bytefmt.MEGABYTE},
		PauseDuration:  duration.Duration{Duration: 5 * time.Second},
		RequestTimeout: duration.Duration{Duration: 1 * time.Minute},
	}

	perfClient, err := perf.NewClient(logger, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "perf client")
	}

	return perfClient, nil
}

func analyzeReports(t *testing.T, reports []*tracker.Report, rssLimit float64) {
	t.Helper()

	sample := &stats.Sample{}

	// take only the second half of observations as we expect memory consumption to be stable here due to MemLimiter work
	reports = reports[len(reports)/2:]

	for _, r := range reports {
		sample.Xs = append(sample.Xs, float64(r.RSS))
	}

	// We can only expect that the memory consumption wouldn't be greater than an
	// upper consumption limit (RSSLimit), but we cannot predict the exact value
	// because of possible existence of a SWAP partition.
	actualRSS := sample.Mean()

	// but since this is a soft limit, we allow small exceeding of it
	require.Less(t, actualRSS, 1.10*rssLimit)
}

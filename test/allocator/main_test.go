/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package main

import (
	"os"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/newcloudtechnologies/memlimiter/test/allocator/app"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestMain(m *testing.M) {
	exitVal := m.Run()
	os.Exit(exitVal)
}

func newLogger() logr.Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	zapLog, err := config.Build()
	if err != nil {
		panic(err)
	}

	return zapr.NewLogger(zapLog)
}

func TestIntegration(t *testing.T) {
	logger := newLogger()

	factory := app.NewFactory(logger)

	srv, err := factory.MakeServerFromConfig("./test/allocator/server/config.json")
	require.NoError(t, err)

	perfClient, err := factory.MakePerfClientFromConfig("./test/allocator/perf/config.json")
	require.NoError(t, err)

	go func() {
		err := srv.Run()
		require.NoError(t, err)
	}()

	time.Sleep(100 * time.Millisecond) // let server start

	err = perfClient.Run()
	require.NoError(t, err)

	perfClient.Quit()

	srv.Quit()
}

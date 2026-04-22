/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2026.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package backpressure

import (
	"github.com/newcloudtechnologies/memlimiter/stats"
	"github.com/stretchr/testify/mock"
)

var _ Operator = (*OperatorMock)(nil)

type OperatorMock struct {
	Operator
	mock.Mock
}

func (m *OperatorMock) SetControlParameters(value *stats.ControlParameters) error {
	args := m.Called(value)

	return args.Error(0)
}

func (m *OperatorMock) AllowRequest() bool {
	args := m.Called()

	return args.Bool(0)
}

func (m *OperatorMock) GetStats() (*stats.BackpressureStats, error) {
	args := m.Called()

	raw := args.Get(0)
	if raw == nil {
		return nil, args.Error(1)
	}

	//nolint:forcetypeassert // Mocked method.
	return raw.(*stats.BackpressureStats), args.Error(1)
}

func (m *OperatorMock) Quit() {
	m.Called()
}

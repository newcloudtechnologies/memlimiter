/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
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

/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package utils

import (
	"github.com/stretchr/testify/mock"
)

type ApplicationTerminatorMock struct {
	mock.Mock
}

func (m *ApplicationTerminatorMock) Terminate(fatalErr error) {
	// TODO implement me
	panic("implement me")
}

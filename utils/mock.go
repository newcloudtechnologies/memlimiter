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

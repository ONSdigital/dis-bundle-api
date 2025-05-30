package mocks

import (
	"net/http"
)

type CheckPermissionFunc func(handler http.HandlerFunc) http.HandlerFunc

type AuthHandlerMock struct {
	Required *PermissionCheckCalls
}

type PermissionCheckCalls struct {
	Calls int
}

func NewAuthHandlerMock() *AuthHandlerMock {
	return &AuthHandlerMock{
		Required: &PermissionCheckCalls{
			Calls: 0,
		},
	}
}

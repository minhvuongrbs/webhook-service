package webhook

import "github.com/pkg/errors"

var (
	ErrRepositoryNotFound = errors.New("repository not found")
	ErrEventNotRegistered = errors.New("event not registered")
)

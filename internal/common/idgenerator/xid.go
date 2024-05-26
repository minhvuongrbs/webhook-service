package idgenerator

import (
	"github.com/rs/xid"
)

// NewId generate unique id with 20 characters
// example: cp9irs04uq1tockiom9g
func NewId() string {
	return xid.New().String()
}

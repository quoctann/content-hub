package domain

import "errors"

// ErrNotFound is returned by repositories when the target row does not exist
// (or is soft-deleted). Callers should match it with errors.Is.
var ErrNotFound = errors.New("not found")

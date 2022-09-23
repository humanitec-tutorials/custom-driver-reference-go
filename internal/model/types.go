//go:generate mockgen -destination mocks/databaser.go humanitec.io/go-service-template/internal/model Databaser

package model

import (
	"context"
	"errors"
)

// ErrNotFound indicates that the resource could not be found.
var ErrNotFound = errors.New("not found")

// ErrConflict indicates that the resource could not be created/updated.
var ErrConflict = errors.New("conflict")

// Databaser provides an interface which can be used to mock the model
type Databaser interface {
	Connect(ctx context.Context, retries int) error
	RunMigrations() error

	SelectEntity(ctx context.Context, id string) (*Entity, error)
}

package model

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Entity stores Entity details.
type Entity struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Status    string     `json:"status"`
	CreatedAt *time.Time `json:"created_at"`
	CreatedBy string     `json:"created_by"`
}

// Value implements the driver.Valuer interface for Entity.
func (entity Entity) Value() (driver.Value, error) {
	return json.Marshal(entity)
}

// Scan implements the sql.Scanner interface for Entity.
func (entity *Entity) Scan(value interface{}) error {
	if bytes, ok := value.([]byte); ok {
		return json.Unmarshal(bytes, &entity)
	}
	return errors.New("type assertion to []byte failed")
}

// SelectEntity loads an Entity details.
// Returns ErrNotFound if the Entity does not exist.
func (db *databaser) SelectEntity(ctx context.Context, id string) (*Entity, error) {
	var qry = `
SELECT
	jsonb_build_object(
		'id', entities.id,
		'name', entities.name,
		'status', entities.status,
		'created_at', to_char(entities.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
		'created_by', entities.created_by
	)
FROM entities
WHERE entities.id = $1`

	row := db.QueryRowContext(ctx, qry, id)

	var res Entity
	if err := row.Scan(&res); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("reading entity: %w", err)
	}

	return &res, nil
}

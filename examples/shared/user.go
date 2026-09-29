// Package: shared
// Description: Shared examples structures
// Author: John Doe
// Version: 1.0.0
package shared

import (
	"time"

	"github.com/google/uuid"
)

// User simple user structure
type User struct {
	// ID unique user ID
	ID uuid.UUID `json:"id" example:"ebdb5d69-2f5c-4732-9daa-e1d2359122df" binding:"required"`
	// Name user name
	Name string `json:"name" example:"user" binding:"required"`
	// Mail user's e-mail
	Mail string `json:"mail" example:"test@mail.ru" binding:"required"`
	// Birth user's birth date
	Birth time.Time `json:"birth_date" example:"2026-01-01" binding:"required"`
}

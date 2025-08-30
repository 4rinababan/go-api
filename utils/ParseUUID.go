package utils

import (
	"github.com/google/uuid"
)

func ParseUUID(id string) uuid.UUID {
	u, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil
	}
	return u
}

package uuid

import (
	guuid "github.com/google/uuid"
)

func NewV7String() string {
	id, err := guuid.NewV7()
	if err != nil {
		return guuid.New().String()
	}
	return id.String()
}

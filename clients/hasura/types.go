package hasura

import (
	"encoding/json"

	"github.com/google/uuid"
)

type UUID uuid.UUID

func (u UUID) GetGraphQLType() string { return "uuid" }

func (u UUID) MarshalJSON() ([]byte, error) {
	s := uuid.UUID(u).String()
	return json.Marshal(s)
}

func (u UUID) MarshalText() ([]byte, error) {
	return []byte(uuid.UUID(u).String()), nil
}

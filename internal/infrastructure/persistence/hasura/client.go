package hasura

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	graphql "github.com/hasura/go-graphql-client"
	"go/food-recipe/internal/config"
)

type Client struct {
	GQL *graphql.Client
}

func NewClient(cfg config.HasuraConfig) *Client {
	if cfg.URL == "" {
		return &Client{GQL: nil}
	}

	c := graphql.NewClient(cfg.URL, &http.Client{Timeout: 30 * time.Second}).
		WithRequestModifier(func(req *http.Request) {
			if cfg.AdminSecret != "" {
				req.Header.Set("x-hasura-admin-secret", cfg.AdminSecret)
			}
		})

	return &Client{GQL: c}
}

// GQLUUID represents a UUID scalar type for Hasura GraphQL operations.
type GQLUUID uuid.UUID

func (u GQLUUID) GetGraphQLType() string { return "uuid" }

func (u GQLUUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(uuid.UUID(u).String())
}

func (u GQLUUID) MarshalText() ([]byte, error) {
	return []byte(uuid.UUID(u).String()), nil
}

func (u *GQLUUID) UnmarshalText(text []byte) error {
	parsed, err := uuid.Parse(string(text))
	if err != nil {
		return err
	}
	*u = GQLUUID(parsed)
	return nil
}

func (u *GQLUUID) UnmarshalJSON(data []byte) error {
	str := string(data)
	if str == "null" || str == `""` || str == "" {
		*u = GQLUUID(uuid.Nil)
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := uuid.Parse(s)
	if err != nil {
		return err
	}
	*u = GQLUUID(parsed)
	return nil
}

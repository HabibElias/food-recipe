package schemas

import "github.com/google/uuid"

type User struct {
	ID        uuid.UUID `graphql:"id"`
	Email     string    `graphql:"email"`
	Password  string    `graphql:"password"`
	UserName  string    `graphql:"username"`
	FirstName string    `graphql:"FirstName"`
	LastName  string    `graphql:"LastName"`
	AvatarUrl string    `graphql:"avatar_url"`
}

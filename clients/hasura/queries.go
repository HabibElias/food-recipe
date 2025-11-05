package hasura

import "go/food-recipe/clients/hasura/schemas"

var GetUsersQuery struct {
	User []schemas.User `graphql:"users"`
}

var GetUserByIDQuery struct {
	User schemas.User `graphql:"user_by_pk(id: $id)"`
}

var GetUserByEmailQuery struct {
	User []schemas.User `graphql:"user(where: {email: {_eq: $email}})"`
}

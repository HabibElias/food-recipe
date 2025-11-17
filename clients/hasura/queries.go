package hasura

import "go/food-recipe/clients/hasura/schemas"

var GetUsersQuery struct {
	User []schemas.User `graphql:"user"`
}

var GetUserByIDQuery struct {
	User schemas.User `graphql:"user_by_pk(id: $id)"`
}

var GetUserByEmailQuery struct {
	User []schemas.User `graphql:"user(where: {email: {_eq: $email}})"`
}

var GetUserByUsernameQuery struct {
	User []schemas.User `graphql:"user(where: {username: {_eq: $username}})"`
}

var GetRecipeCreatorFromLikeIdQuery struct {
	Like struct {
		ID     string `graphql:"id"`
		Recipe struct {
			UserID string `graphql:"user_id"`
		} `graphql:"recipe"`
		User struct {
			ID       string `graphql:"id"`
			Username string `graphql:"username"`
		} `graphql:"user"`
	} `graphql:"recipe_like_by_pk(id: $id)"`
}

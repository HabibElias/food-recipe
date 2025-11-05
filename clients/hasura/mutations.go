package hasura

var InsertUser struct {
	InsertUsers struct {
		Returning []struct {
			ID    string
			Name  string
			Email string
		}
	} `graphql:"insert_users(objects: {email: $email, password: $password, username: $username})"`
}

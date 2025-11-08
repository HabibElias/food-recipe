package hasura

var InsertUser struct {
	InsertUsers struct {
		Returning []struct {
			ID    string
			Email string
		}
	} `graphql:"insert_user(objects: {email: $email, password: $password, username: $username, FirstName: $first_name, LastName: $last_name})"`
}

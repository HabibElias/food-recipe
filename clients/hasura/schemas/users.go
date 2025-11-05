package schemas

type User struct {
	ID       string    `graphql:"id"`
	Email    string `graphql:"email"`
	Password string `graphql:"password"`
	UserName     string `graphql:"username"`
}

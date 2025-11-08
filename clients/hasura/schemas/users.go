package schemas

type User struct {
	ID        string `graphql:"id"`
	Email     string `graphql:"email"`
	Password  string `graphql:"password"`
	UserName  string `graphql:"username"`
	FirstName string `graphql:"FirstName"`
	LastName  string `graphql:"LastName"`
	AvatarUrl string `graphql:"avatar_url"`
}

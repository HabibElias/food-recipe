package entities

type RecipeImage struct {
	ID          string
	RecipeID    int
	ImageURL    string
	IsThumbnail bool
}

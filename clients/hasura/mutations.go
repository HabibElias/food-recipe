package hasura

import "github.com/google/uuid"

var InsertUser struct {
	InsertUsers struct {
		Returning []struct {
			ID    string
			Email string
		}
	} `graphql:"insert_user(objects: {email: $email, password: $password, username: $username, FirstName: $first_name, LastName: $last_name})"`
}

var InsertPayment struct {
	InsertPayment struct {
		Returning []struct {
			Id string `graphql:"id"`
		}
	} `graphql:"insert_payment(objects: {user_id: $user_id, recipe_id: $recipe_id, tx_ref: $tx_ref, amount: $amount, chapa_checkout_url: $checkout_url, currency: $currency, status: $status})"`
}

var InsertRecipeImage struct {
	InsertRecipeImage struct {
		Returning []struct {
			ID string `graphql:"id"`
		}
	} `graphql:"insert_recipe_image(objects: {img_url: $img_url, recipe_id: $recipe_id, is_thumbnail: $is_thumbnail})"`
}

var CreateNotificationMutation struct {
	InsertNotifications struct {
		Returning []struct {
			ID uuid.UUID `graphql:"id"`
		}
	} `graphql:"insert_notifications(objects: {user_id: $user_id, message: $message, initiator_id: $initiator_id})"`
}

var UpdateUserAvatar struct {
	UpdateUsers struct {
		Returning []struct {
			ID string `graphql:"id"`
		}
	} `graphql:"update_user(where: {id: {_eq: $id}}, _set: {avatar_url: $avatar_url})"`
}

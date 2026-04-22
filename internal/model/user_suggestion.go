package model

//easyjson:json
type UserSuggestion struct {
	ID    string `json:"user_id"`
	Login string `json:"login"`
}

//easyjson:json
type UserSuggestions struct {
	Items []UserSuggestion `json:"items"`
}

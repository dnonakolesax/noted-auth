package model

type User struct { //nolint:recvcheck // autogen issues
	Login     string `json:"login"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

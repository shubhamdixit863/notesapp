package models

type User struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
	Id       int    `json:"id,omitempty"`
}

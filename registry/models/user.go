package models

type User struct {
	ID          string `json:"_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type LoginResponse struct {
	Token string `json:"token"` // the Bearer token
	OK    bool   `json:"ok"`
	ID    string `json:"id"`
}

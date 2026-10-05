package models

type User struct {
	ID           int64
	PasswordHash string
	Role         string
}

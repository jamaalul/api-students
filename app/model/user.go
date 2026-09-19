package model

import "time"

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Role      Role      `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

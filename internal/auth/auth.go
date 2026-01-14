package auth

import "github.com/golang-jwt/jwt/v5"

type AuthContextKey struct{}

type Claims struct {
	Id       int32  `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

var JWTKey = []byte("my_secret_key")

const (
	RoleStaff = "staff"
	RoleUser  = "user"
)

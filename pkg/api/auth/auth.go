package client

import (
	"encoding/json"
	"fmt"
	"main/internal/endpoint/auth"
	"net/http"
)

func UserLogin(user User) (*TokenPair, error) {
	req, err := http.NewRequest(http.MethodPost, "localhost:8090/api"+"/auth/login", nil)
	if err != nil {
		fmt.Printf("unable to create request: %v\n", err)
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("unable to send request: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close() // nolint

	refreshTokenCookie, err := resp.Request.Cookie("refreshToken")
	if err != nil {
		fmt.Printf("no refresh token in cookie\n")
		// return nil, err
	}

	var res auth.LoginResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		fmt.Printf("unable to parse response: %v\n", err)
		return nil, err
	}

	refreshTokenStr := ""
	if refreshTokenCookie == nil {
		refreshTokenStr = "test"
	} else {
		refreshTokenStr = refreshTokenCookie.String()
	}

	return &TokenPair{Access: res.Access, Refresh: refreshTokenStr}, nil
}

type User struct {
	Username string
	Password string
}

type TokenPair struct {
	Refresh string
	Access  string
}

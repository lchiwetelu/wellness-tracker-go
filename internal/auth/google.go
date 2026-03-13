package auth

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func NewGoogleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  os.Getenv("BASE_URL") + "/auth/google/callback",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}


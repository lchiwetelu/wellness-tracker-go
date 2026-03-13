package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"wellness_tracker/internal/auth"
	"wellness_tracker/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

type GoogleUser struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func (a *AuthHandler) Login(c *gin.Context) {
	// 1. Generate random state to prevent CSRF
	state, _ := generateState()

	// 2. Store it in a cookie (so we can verify it later in callback)
	c.SetCookie("oauth_state", state, 300, "/", "", false, true) // expires in 5 min

	// 3. Build the Google login URL
	config := auth.NewGoogleOAuthConfig()
	url := config.AuthCodeURL(state)

	// 4. Redirect browser to Google
	c.Redirect(http.StatusTemporaryRedirect, url)

}

func generateState() (string, error) {
	b := make([]byte, 16) // 16 bytes = 128 bits
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (a *AuthHandler) GoogleCallback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")

	if state == "" || code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing state or code"})
		return
	}

	if !verifyState(c, state) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		return
	}

	token, err := exchangeCodeForToken(c, code)
	if err != nil {
		log.Printf("failed to exchange token. err %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	gUser, err := fetchGoogleUser(c, token)
	if err != nil {
		log.Printf("failed to fetch google user. err %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	user, err := a.findOrCreateUser(gUser)
	if err != nil {
		log.Printf("failed to find or create user. err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// 1. Generate JWT
	jwtToken, err := auth.GenerateJWT(user.ID, user.Email, user.AuthProvider)
	if err != nil {

		log.Printf("failed to generate JWT for user %d: %v", user.ID, err)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// 2. Set HttpOnly cookie
	c.SetCookie(
		"jwt",       // name
		jwtToken,    // value
		3600*24,     // 1 day in seconds
		"/",         // path
		"localhost", // domain (change to your frontend domain in production)
		true,        // secure (HTTPS only in production)
		true,        // HttpOnly
	)

	frontendUrl := os.Getenv("FRONTEND_URL")

	// 3. Redirect to frontend dashboard
	c.Redirect(http.StatusTemporaryRedirect, frontendUrl)
}

func verifyState(c *gin.Context, state string) bool {
	cookieState, err := c.Cookie("oauth_state")
	if err != nil || cookieState != state {
		return false
	}
	return true
}

// exchangeCodeForToken exchanges OAuth code for access token
func exchangeCodeForToken(c *gin.Context, code string) (*oauth2.Token, error) {
	return auth.NewGoogleOAuthConfig().Exchange(c.Request.Context(), code)
}

func fetchGoogleUser(c *gin.Context, token *oauth2.Token) (*GoogleUser, error) {
	client := auth.NewGoogleOAuthConfig().Client(c.Request.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil || resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get user info")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var gUser GoogleUser
	if err := json.Unmarshal(body, &gUser); err != nil {
		return nil, fmt.Errorf("failed to parse user info")
	}

	return &gUser, nil
}

func (a *AuthHandler) findOrCreateUser(gUser *GoogleUser) (*models.User, error) {
	var user models.User
	err := a.db.Where("google_id = ?", gUser.ID).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		user = models.User{
			Email:        gUser.Email,
			Name:         gUser.Name,
			Picture:      gUser.Picture,
			GoogleID:     gUser.ID,
			AuthProvider: "google",
		}
		if err := a.db.Create(&user).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return &user, nil
}

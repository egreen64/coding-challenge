package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/egreen64/codingchallenge/config"
)

// A private key for context that only this package can access. This is important
// to prevent collisions between different context uses
type contextKey struct {
	name string
}

var (
	signingKey      = []byte("secret")
	authTokenCtxKey = &contextKey{"auth-token"}
)

type customClaims struct {
	Username string `json:"username"`
	Password string `json:"passowrd"`
	jwt.StandardClaims
}

//CreateJWT function
func CreateJWT(username string, password string, expirationDuration int) (string, error) {
	// Create the Claims
	claims := customClaims{
		username,
		password,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Duration(expirationDuration) * time.Minute).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}

	return ss, nil
}

//ValidateJWT function
func ValidateJWT(tokenString string, username string, password string) (bool, error) {
	token, err := jwt.ParseWithClaims(tokenString, &customClaims{}, func(token *jwt.Token) (interface{}, error) {
		return signingKey, nil
	})

	if err != nil {
		return false, err
	}

	claims, ok := token.Claims.(*customClaims)
	if !ok || !token.Valid {
		return false, errors.New("invalid token")
	}

	if claims.Username != username || claims.Password != password {
		return false, errors.New("invalid token")
	}
	return true, nil
}

//Middleware decodes the share session cookie and packs the session into context
func Middleware(config *config.File) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bearerToken := r.Header.Get("Authorization")

			// Allow unauthenticated users in
			if bearerToken == "" {
				next.ServeHTTP(w, r)
				return
			}

			// put it in context
			ctx := context.WithValue(r.Context(), authTokenCtxKey, bearerToken)

			// and call the next with our new context
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

//GetContextToken gets the auth token from the request context
func GetContextToken(ctx context.Context) string {
	tokenString, _ := ctx.Value(authTokenCtxKey).(string)
	return tokenString
}

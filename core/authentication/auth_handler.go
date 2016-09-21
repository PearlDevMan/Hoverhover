package authentication

import (
	"fmt"
	"net/http"
	log "github.com/Sirupsen/logrus"

	"github.com/SpectoLabs/hoverfly/core/authentication/backends"
	jwt "github.com/dgrijalva/jwt-go"
	"encoding/json"
)

type AuthHandler struct {
	AB                 backends.Authentication
	SecretKey          []byte
	JWTExpirationDelta int
	Enabled            bool
}

func (a *AuthHandler) RequireTokenAuthentication(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	// if auth is disabled - do not check token
	if !a.Enabled {
		next(w, req)
		return
	}

	authBackend := InitJWTAuthenticationBackend(a.AB, a.SecretKey, a.JWTExpirationDelta)

	token, err := jwt.ParseFromRequest(req, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		} else {
			return authBackend.SecretKey, nil
		}
	})

	if err == nil && token.Valid && !authBackend.IsInBlacklist(req.Header.Get("Authorization")) {
		next(w, req)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

type AllUsersResponse struct {
	Users []backends.User `json:"users"`
}

func (a *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !a.Enabled {
		w.WriteHeader(http.StatusOK)
		// returning dummy token
		token := `{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkR1bW15IHRva2VuIiwiYWRtaW4iOnRydWV9.sKfJparPo3LUmkYoGboBjVfOV3K1qWKUzqx9XFDEsAs"}`
		w.Write([]byte(token))
		return
	}
	requestUser := new(backends.User)
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&requestUser)

	responseStatus, token := Login(requestUser, a.AB, a.SecretKey, a.JWTExpirationDelta)

	w.WriteHeader(responseStatus)
	w.Write(token)
}

func (a *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	w.Header().Set("Content-Type", "application/json")

	requestUser := new(backends.User)
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&requestUser)

	w.Write(RefreshToken(requestUser, a.AB, a.SecretKey, a.JWTExpirationDelta))
}

func (a *AuthHandler) Logout(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	w.Header().Set("Content-Type", "application/json")
	if !a.Enabled {
		w.WriteHeader(http.StatusOK)
		return
	}

	err := Logout(r, a.AB, a.SecretKey, a.JWTExpirationDelta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

// GetAllUsersHandler - returns a list of all users
func (a *AuthHandler) GetAllUsersHandler(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	users, err := a.AB.GetAllUsers()

	if err == nil {

		w.Header().Set("Content-Type", "application/json")

		var response AllUsersResponse
		response.Users = users
		b, err := json.Marshal(response)

		if err != nil {
			log.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.Write(b)
			return
		}
	} else {
		log.WithFields(log.Fields{
			"Error": err.Error(),
		}).Error("Failed to get data from authentication backend!")

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(500)
		return
	}
}

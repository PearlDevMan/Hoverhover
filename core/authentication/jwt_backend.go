package authentication

import (
	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"bytes"
	"encoding/gob"
	"time"

	"github.com/SpectoLabs/hoverfly/core/authentication/backends"
)

type JWTAuthenticationBackend struct {
	SecretKey          []byte
	JWTExpirationDelta int
	AuthBackend        backends.Authentication
}

const (
	expireOffset = 3600
)

//Token - container for jwt.Token for encoding
type Token struct {
	Token *jwt.Token
}

func (t *Token) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(t)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func InitJWTAuthenticationBackend(ab backends.Authentication, secret []byte, exp int) *JWTAuthenticationBackend {
	return &JWTAuthenticationBackend{
		SecretKey:          secret,
		AuthBackend:        ab,
		JWTExpirationDelta: exp,
	}
}

func (backend *JWTAuthenticationBackend) GenerateToken(userUUID, username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS512)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(backend.JWTExpirationDelta)).Unix()
	claims["iat"] = time.Now().Unix()
	claims["username"] = username
	claims["sub"] = userUUID
	tokenString, err := token.SignedString(backend.SecretKey)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("got error while generating JWT token")
		return "", err
	}
	return tokenString, nil
}

func (backend *JWTAuthenticationBackend) Authenticate(user *backends.User) bool {
	dbUser, err := backend.AuthBackend.GetUser(user.Username)
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err.Error(),
			"username": user.Username,
		}).Error("error while getting user")
		return false
	}

	// user does not exist
	if dbUser == nil {
		log.WithFields(log.Fields{
			"username": user.Username,
		}).Warn("user does not exist")
		return false
	}

	return user.Username == dbUser.Username && bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password)) == nil
}

func (backend *JWTAuthenticationBackend) getTokenRemainingValidity(timestamp interface{}) int {
	if validity, ok := timestamp.(float64); ok {
		tm := time.Unix(int64(validity), 0)
		remainer := tm.Sub(time.Now())
		if remainer > 0 {
			return int(remainer.Seconds() + expireOffset)
		}
	}
	return expireOffset
}

func (backend *JWTAuthenticationBackend) Logout(tokenString string) error {
	return backend.AuthBackend.InvalidateToken(tokenString)
}

func (backend *JWTAuthenticationBackend) IsInBlacklist(token string) bool {
	blacklisted, _ := backend.AuthBackend.IsTokenBlacklisted(token)
	return blacklisted
}

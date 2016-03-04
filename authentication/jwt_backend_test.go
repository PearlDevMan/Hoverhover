package authentication

import (
    "testing"
    "os"
    
	"github.com/SpectoLabs/hoverfly/authentication/backends"
)


// TestMain prepares database for testing and then performs a cleanup
func TestMain(m *testing.M) {
	setup()
	retCode := m.Run()
	// delete test database
	teardown()
	// call with result of m.Run()
	os.Exit(retCode)
}

func TestGenerateToken(t *testing.T) {
       ab := backends.NewBoltDBAuthBackend(TestDB, []byte(backends.TokenBucketName), []byte(backends.UserBucketName))
       jwtBackend := InitJWTAuthenticationBackend(ab, []byte("verysecret"), 100)
       
      token, err := jwtBackend.GenerateToken("userUUIDhereVeryLong", "userx")
      expect(t, err, nil)
      expect(t, len(token) > 0, true)
}
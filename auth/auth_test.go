package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	//Disable logging
	//log.SetOutput(ioutil.Discard)

	t.Run("create_jwt_success", func(t *testing.T) {

		token, err := CreateJWT("secureworks", "supersecret", 15)
		require.NotEqual(t, "", token)
		require.Equal(t, nil, err)
	})

	t.Run("validate_token_success", func(t *testing.T) {

		token, err := CreateJWT("secureworks", "supersecret", 15)
		require.NotEqual(t, "", token)
		require.Equal(t, nil, err)

		resp, err := ValidateJWT(token, "secureworks", "supersecret")
		require.Equal(t, true, resp)
		require.Equal(t, nil, err)
	})

	t.Run("validate_token_failure_invalid_username", func(t *testing.T) {

		token, err := CreateJWT("secureworks", "supersecret", 15)
		require.NotEqual(t, "", token)
		require.Equal(t, nil, err)

		resp, err := ValidateJWT(token, "securework", "supersecret")
		require.Equal(t, false, resp)
		require.NotEqual(t, nil, err)
		require.Equal(t, "invalid token", err.Error())
	})
	t.Run("validate_token_failure_invalid_password", func(t *testing.T) {

		token, err := CreateJWT("secureworks", "supersecret", 15)
		require.NotEqual(t, "", token)
		require.Equal(t, nil, err)

		resp, err := ValidateJWT(token, "secureworks", "supersecrets")
		require.Equal(t, false, resp)
		require.NotEqual(t, nil, err)
		require.Equal(t, "invalid token", err.Error())
	})
	t.Run("validate_token_failure_invalid_token", func(t *testing.T) {

		token, err := CreateJWT("secureworks", "supersecret", 15)
		require.NotEqual(t, "", token)
		require.Equal(t, nil, err)

		token = token[1:]
		resp, err := ValidateJWT(token, "secureworks", "supersecret")
		require.Equal(t, false, resp)
		require.NotEqual(t, nil, err)
		require.Equal(t, "invalid token", err.Error())
	})
}

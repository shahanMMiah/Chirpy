package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestJwtMakeValidate(t *testing.T) {
	cases := []struct {
		InputId     uuid.UUID
		InputSecret string
	}{
		{
			InputId:     uuid.New(),
			InputSecret: "testSecret"},
	}
	for _, c := range cases {

		jwt_token, err := MakeJWT(c.InputId, c.InputSecret, 1*time.Minute)
		if err != nil {
			t.Errorf("error making token: %s", err.Error())

		}

		actualId, err := ValidateJWT(jwt_token, c.InputSecret)

		if err != nil {
			t.Errorf("error validating token: %s", err.Error())

		}

		if actualId != c.InputId {
			t.Errorf("error id from validated token differs from input %s, %s", actualId.String(), c.InputId.String())
		}

	}

}
func TestJwtInvalidate(t *testing.T) {
	cases := []struct {
		InputId          uuid.UUID
		InputSecret      string
		InputWrongSecret string
	}{
		{
			InputId:          uuid.New(),
			InputSecret:      "testSecret",
			InputWrongSecret: "notSameSecret"},
	}
	for _, c := range cases {

		jwt_token, err := MakeJWT(c.InputId, c.InputSecret, 1*time.Minute)

		if err != nil {
			t.Errorf("error making token: %s", err.Error())

		}

		_, err = ValidateJWT(jwt_token, c.InputWrongSecret)

		if err == nil {
			t.Error("error should not be nil")
			return
		}

		if !strings.Contains(err.Error(), jwt.ErrTokenSignatureInvalid.Error()) {
			t.Errorf("error is %s should be %s", err.Error(), jwt.ErrTokenSignatureInvalid.Error())

		}

	}

}

func TestJwtExpires(t *testing.T) {
	cases := []struct {
		InputId             uuid.UUID
		InputSecret         string
		InputDuration       time.Duration
		InputExpireDuration time.Duration
	}{
		{
			InputId:             uuid.New(),
			InputSecret:         "testSecret",
			InputDuration:       1 * time.Millisecond,
			InputExpireDuration: 2 * time.Millisecond},
	}
	for _, c := range cases {

		jwt_token, err := MakeJWT(c.InputId, c.InputSecret, c.InputDuration)

		if err != nil {
			t.Errorf("error making token: %s", err.Error())

		}

		timer := time.NewTicker(c.InputExpireDuration)

		<-timer.C

		_, err = ValidateJWT(jwt_token, c.InputSecret)

		if err == nil {
			t.Error("error should not be nil")
			return
		}

		if !strings.Contains(err.Error(), jwt.ErrTokenExpired.Error()) {
			t.Errorf("error should be:'%s'\n but error is: '%s'", jwt.ErrTokenExpired.Error(), err.Error())

		}

	}
}

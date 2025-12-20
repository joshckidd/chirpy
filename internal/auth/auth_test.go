package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCorrectPassword(t *testing.T) {
	hashedPassword, err := HashPassword("1234")
	if err != nil {
		t.Errorf("Error hashing password: %v", err.Error())
	}
	res, err := CheckPasswordHash("1234", hashedPassword)
	if res == false || err != nil {
		t.Errorf("Want true, got %v. Error: %v", res, err.Error())
	}
}

func TestIncorrectPassword(t *testing.T) {
	hashedPassword, err := HashPassword("1234")
	if err != nil {
		t.Errorf("Error hashing password: %v", err.Error())
	}
	res, err := CheckPasswordHash("1235", hashedPassword)
	if res == true || err != nil {
		t.Errorf("Want false, got %v. Error: %v", res, err.Error())
	}
}

func TestCorrectJWT(t *testing.T) {
	id := uuid.New()
	tokenString, err := MakeJWT(id, "1234", time.Hour)
	if err != nil {
		t.Errorf("Error making JWT: %v", err.Error())
	}
	res, err := ValidateJWT(tokenString, "1234")
	if res != id || err != nil {
		t.Errorf("Want %v, got %v. Error: %v", id.String(), res.String(), err.Error())
	}
}

func TestJWTIncorrectSecret(t *testing.T) {
	id := uuid.New()
	tokenString, err := MakeJWT(id, "1234", time.Hour)
	if err != nil {
		t.Errorf("Error making JWT: %v", err.Error())
	}
	res, err := ValidateJWT(tokenString, "1235")
	if err.Error() != "token signature is invalid: signature is invalid" {
		t.Errorf("Want %v, got %v. Error: %v", "token signature is invalid", res.String(), err.Error())
	}
}

func TestGetBearerToken(t *testing.T) {
	h := http.Header{}
	h.Add("Authorization", "Bearer TOKEN_STRING")
	tok, err := GetBearerToken(h)
	if tok != "TOKEN_STRING" || err != nil {
		t.Errorf("Want %v, got %v. Error: %v", "TOKEN_STRING", tok, err.Error())
	}
}

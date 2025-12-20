package auth

import "testing"

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

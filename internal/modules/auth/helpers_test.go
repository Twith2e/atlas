package auth

import (
	appErr "atlas/internal/errors"
	"fmt"
	"testing"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{"valid password", "Pa55word!", nil},
		{"too short", "P5w!", appErr.ErrInvalidPasswordLength},
		{"exactly 8 chars valid", "Pa5word!", nil},
		{"missing upper", "pa55word!", appErr.ErrInvalidPassword},
		{"missing lower", "PA55WORD!", appErr.ErrInvalidPassword},
		{"missing number", "Password!", appErr.ErrInvalidPassword},
		{"missing special", "Pa55wordd", appErr.ErrInvalidPassword},
		{"disallowed special char only", "Pa55word~", appErr.ErrInvalidPassword},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if err != tt.wantErr {
				t.Errorf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func ExampleValidatePassword() {
	password := "Pa55word!"
	err := ValidatePassword(password)
	fmt.Printf("%v\n", err)
	// Output: <nil>
}

func TestComparePasswords(t *testing.T) {
	tests := []struct {
		name           string
		password       string
		hashedPassword string
		wantErr        error
	}{
		{"password match", "Pa55word!", "$2a$10$Vi5vrJZgfvzYCnOB3BYm7./ebQpK0RFQK9Lh1YyIIY37gW3vPYAxS", nil},
		{"password does not match", "Pa55word!", "", appErr.ErrPasswordMismatch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ComparePasswords(tt.hashedPassword, tt.password)
			if err != tt.wantErr {
				t.Errorf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

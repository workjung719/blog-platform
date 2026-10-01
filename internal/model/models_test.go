package model

import "testing"

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"валидный", "user@example.com", false},
		{"валидный с поддоменом", "a.b@c.co.uk", false},
		{"без собаки", "invalidemail", true},
		{"домен без точки", "a@b", true},
		{"пустой", "", true},
		{"точка в начале домена", "a@.com", true},
		{"двойная точка в домене", "a@b..com", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail(%q) err = %v, wantErr %v", tt.email, err, tt.wantErr)
			}
		})
	}
}

func TestRegisterRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     RegisterRequest
		wantErr bool
	}{
		{"корректный", RegisterRequest{Email: "u@e.com", Username: "gooduser", Password: "longpass123"}, false},
		{"короткий пароль", RegisterRequest{Email: "u@e.com", Username: "gooduser", Password: "short"}, true},
		{"короткий username", RegisterRequest{Email: "u@e.com", Username: "ab", Password: "longpass123"}, true},
		{"длинный username", RegisterRequest{Email: "u@e.com", Username: "thisisaverylongusernameoverfiftycharsjusttotestvalidationlogic", Password: "longpass123"}, true},
		{"невалидный email", RegisterRequest{Email: "bad", Username: "gooduser", Password: "longpass123"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

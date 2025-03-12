package auth

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestNewAuth(t *testing.T) {
	tests := []struct {
		name    string
		args    AuthConfig
		want    *Auth
		wantErr bool
	}{
		{
			name: "Test case 1",
			args: AuthConfig{
				Username:      "test",
				Password:      "testtest",
				SecretKey:     "thisisaverylongsecretkeythatisatleast32characterslong",
				CookieName:    "testcookie",
				TokenValidity: 60,
			},
			want: &Auth{
				Username:      "test",
				Password:      "testtest",
				SecretKey:     "thisisaverylongsecretkeythatisatleast32characterslong",
				CookieName:    "testcookie",
				TokenValidity: 60,
			},
			wantErr: false,
		},
		{
			name: "Test case 2",
			args: AuthConfig{
				Username:      "test",
				Password:      "testtest",
				SecretKey:     "thisisaverysmallsecretkey",
				CookieName:    "testcookie",
				TokenValidity: 60,
			},
			want:    nil,
			wantErr: true,
		},

		{
			name: "Test case 3",
			args: AuthConfig{
				Username:      "123",
				Password:      "testtest",
				SecretKey:     "thisisaverylongsecretkeythatisatleast32characterslong",
				CookieName:    "testcookie",
				TokenValidity: 60,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test case 4",
			args: AuthConfig{
				Username:      "teset",
				Password:      "test",
				SecretKey:     "thisisaverylongsecretkeythatisatleast32characterslong",
				CookieName:    "testcookie",
				TokenValidity: 60,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test case 5",
			args: AuthConfig{
				Username:      "teset",
				Password:      "testtest",
				SecretKey:     "thisisaverylongsecretkeythatisatleast32characterslong",
				CookieName:    "testcookie",
				TokenValidity: 10,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test case 6",
			args: AuthConfig{
				Username:      "teset",
				Password:      "testtest",
				SecretKey:     "thisisaverylongsecretkeythatisatleast32characterslong",
				CookieName:    "test",
				TokenValidity: 60,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test case 7",
			args: AuthConfig{
				Username:      "teset",
				Password:      "testtest",
				SecretKey:     "thisisaverylongsecretkeythatisatleast32characterslong",
				CookieName:    "test:cookie",
				TokenValidity: 60,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test case 8",
			args: AuthConfig{
				Username:      "tese:t",
				Password:      "testtest",
				SecretKey:     "thisisaverylongsecretkeythatisatleast32characterslong",
				CookieName:    "testcookie",
				TokenValidity: 60,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test case 9",
			args: AuthConfig{
				Username:      "testtest",
				Password:      "testtest",
				SecretKey:     "thisisaverylongsecretkeythatisatleast32characterslong",
				CookieName:    "testcookie",
				TokenValidity: 60,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAuth(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAuth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuth_GenerateAndValidateToken(t *testing.T) {
	secret := "thisisaverylongsecretkeythatisatleast32characterslong"

	type fields struct {
		Username      string
		Password      string
		SecretKey     string
		TokenValidity int64
		CookieName    string
	}

	tests := []struct {
		fields fields
		name   string
	}{
		{
			fields: fields{
				Username:      "test",
				Password:      "testtest",
				SecretKey:     secret,
				TokenValidity: 60,
				CookieName:    "testcookie",
			},
			name: "Test case 1",
		},

		{
			fields: fields{
				Username:      "teste123",
				Password:      "testtest",
				SecretKey:     secret,
				TokenValidity: 60,
				CookieName:    "testcookie",
			},
			name: "Test case 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, _ := NewAuth(AuthConfig{
				Username:      tt.fields.Username,
				Password:      tt.fields.Password,
				SecretKey:     tt.fields.SecretKey,
				TokenValidity: tt.fields.TokenValidity,
				CookieName:    tt.fields.CookieName,
			})

			token := a.GenerateToken(tt.fields.Username, time.Now().Unix()+tt.fields.TokenValidity)

			valid, err := a.ValidateToken(token)
			if err != nil {
				t.Errorf("ValidateToken() error = %v", err)
				return
			}

			if !valid {
				t.Errorf("ValidateToken() = %v, want %v", valid, true)
			}

			username := strings.Split(token, ":")[0]

			if username != tt.fields.Username {
				t.Errorf("ValidateToken() = %v, want %v", username, tt.fields.Username)
			}
		})
	}
}

package auth

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestNewAuth(t *testing.T) {
	type args struct {
		username      string
		password      string
		secretKey     string
		cookieName    string
		tokenValidity int64
	}
	tests := []struct {
		name    string
		args    args
		want    *Auth
		wantErr bool
	}{
		{
			name: "Test case 1",
			args: args{
				username:      "test",
				password:      "testtest",
				secretKey:     "thisisaverylongsecretkeythatisatleast32characterslong",
				cookieName:    "testcookie",
				tokenValidity: 60,
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
			args: args{
				username:      "test",
				password:      "testtest",
				secretKey:     "thisisaverysmallsecretkey",
				cookieName:    "testcookie",
				tokenValidity: 60,
			},
			want:    nil,
			wantErr: true,
		},

		{
			name: "Test case 3",
			args: args{
				username:      "123",
				password:      "testtest",
				secretKey:     "thisisaverylongsecretkeythatisatleast32characterslong",
				cookieName:    "testcookie",
				tokenValidity: 60,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test case 4",
			args: args{
				username:      "teset",
				password:      "test",
				secretKey:     "thisisaverylongsecretkeythatisatleast32characterslong",
				cookieName:    "testcookie",
				tokenValidity: 60,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test case 5",
			args: args{
				username:      "teset",
				password:      "testtest",
				secretKey:     "thisisaverylongsecretkeythatisatleast32characterslong",
				cookieName:    "testcookie",
				tokenValidity: 10,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test case 6",
			args: args{
				username:      "teset",
				password:      "testtest",
				secretKey:     "thisisaverylongsecretkeythatisatleast32characterslong",
				cookieName:    "test",
				tokenValidity: 60,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test case 7",
			args: args{
				username:      "teset",
				password:      "testtest",
				secretKey:     "thisisaverylongsecretkeythatisatleast32characterslong",
				cookieName:    "test:cookie",
				tokenValidity: 60,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test case 8",
			args: args{
				username:      "tese:t",
				password:      "testtest",
				secretKey:     "thisisaverylongsecretkeythatisatleast32characterslong",
				cookieName:    "testcookie",
				tokenValidity: 60,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test case 9",
			args: args{
				username:      "testtest",
				password:      "testtest",
				secretKey:     "thisisaverylongsecretkeythatisatleast32characterslong",
				cookieName:    "testcookie",
				tokenValidity: 60,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAuth(tt.args.username, tt.args.password, tt.args.secretKey, tt.args.cookieName, tt.args.tokenValidity)
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
			a := &Auth{
				Username:      tt.fields.Username,
				Password:      tt.fields.Password,
				SecretKey:     tt.fields.SecretKey,
				TokenValidity: tt.fields.TokenValidity,
				CookieName:    tt.fields.CookieName,
			}

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

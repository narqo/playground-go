package easyjson

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestEmail_UnmarshalJSON(t *testing.T) {
	cases := []struct {
		in      string
		email   Email
		wantErr bool
	}{
		{
			``,
			"",
			true,
		},
		{
			`""`,
			"",
			true,
		},
		{
			`"foo"`,
			"",
			true,
		},
		{
			`"root@localhost"`,
			"",
			true,
		},
		{
			`"foo@localhost"`,
			"foo@localhost",
			false,
		},
	}

	for n, tc := range cases {
		t.Run(fmt.Sprintf("case=%d", n), func(t *testing.T) {
			var email Email
			err := json.Unmarshal([]byte(tc.in), &email)
			if err != nil {
				t.Logf("could not unmarshal email %q: %v", tc.in, err)
			}

			if (err != nil) != tc.wantErr {
				t.Errorf("want error %v, got %v", tc.wantErr, err)
			}
			if email != tc.email {
				t.Errorf("want email %q, got %q", tc.email, email)
			}
		})
	}

	t.Run("unmarshal struct", func(t *testing.T) {
		jsonUser := `{"name":"Captain Obvious", "email": "root@localhost"}`

		user := struct {
			Name  string `json:"name"`
			Email Email  `json:"email"`
		}{}

		err := json.Unmarshal([]byte(jsonUser), &user)
		if err != nil {
			t.Logf("could not unmarshal message: %v", err)
			return
		}
		t.Error("want error, got none")
	})
}

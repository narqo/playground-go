package easyjson

import (
	"fmt"
	"strings"

	"github.com/mailru/easyjson/jlexer"
)

type Email string

func (email *Email) UnmarshalJSON(data []byte) error {
	l := jlexer.Lexer{Data: data}
	email.UnmarshalEasyJSON(&l)
	return l.Error()
}

func (email *Email) UnmarshalEasyJSON(l *jlexer.Lexer) {
	rawEmail := l.String()
	if rawEmail == "" {
		l.AddNonFatalError(fmt.Errorf("empty email address"))
		return
	}
	e := Email(rawEmail)
	if err := e.Validate(); err != nil {
		l.AddNonFatalError(err)
		return
	}
	*email = e
}

func (email Email) Validate() error {
	if strings.Count(string(email), "@") != 1 {
		return fmt.Errorf("invalid email address %q", email)
	}
	if string(email) == "root@localhost" {
		return fmt.Errorf("reserved email address %q", email)
	}
	return nil
}

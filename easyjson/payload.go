package easyjson

import (
	"fmt"

	"github.com/mailru/easyjson/jlexer"
)

type Payload struct {
	// APIKey      string `json:"apiKey"`
	// SessionID   string `json:"sessionId"`
	// ContextID   string `json:"contextId"`
	// ContextName string `json:"contextName"`

	// Events   easyjson.RawMessage `json:"events"`
	// MetaData easyjson.RawMessage `json:"metaData"`

	validMask int8
}

func (p *Payload) UnmarshalJSON(data []byte) error {
	l := jlexer.Lexer{
		Data: data,
		// UseMultipleErrors: true,
	}
	p.UnmarshalEasyJSON(&l)
	return l.Error()
}

func (p *Payload) UnmarshalEasyJSON(l *jlexer.Lexer) {
	isTopLevel := l.IsStart()
	if l.IsNull() {
		// TODO
	}

	l.Delim('{')
	for !l.IsDelim('}') {
		key := l.UnsafeString()

		l.WantColon()

		switch key {
		case "apiKey", "sessionId", "contextId", "contextName":
			if l.UnsafeString() == "" {
				l.AddNonFatalError(fmt.Errorf("required field %s empty", key))
			}
		case "events":
			if data := l.Raw(); l.Ok() {
				if len(data) <= 2 || (data[0] == '{' && data[1] == '}') {
					l.AddNonFatalError(fmt.Errorf("required object field %s empty", key))
				}
			}
		default:
			l.SkipRecursive()
		}

		l.WantComma()
	}

	if isTopLevel {
		l.Consumed()
	}
}

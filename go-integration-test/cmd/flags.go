package main

import (
	"strings"
)

type stringsValue []string

func (ss stringsValue) Get() interface{} {
	return ss
}

func (ss *stringsValue) Set(s string) error {
	for _, v := range strings.Split(s, ",") {
		*ss = append(*ss, strings.TrimSpace(v))
	}
	return nil
}

func (ss stringsValue) String() string {
	if ss == nil {
		return ""
	}
	return strings.Join(ss, ",")
}


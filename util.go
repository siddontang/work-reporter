package main

import (
	"strings"
)

func escaperValue(value string) string {
	value = strings.ToValidUTF8(value, "")
	value = strings.ReplaceAll(value, "&", "&amp;")
	return value
}

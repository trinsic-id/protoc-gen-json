package main

import (
	"strings"
)

func StripStartingPeriod(str string) string {
	return strings.TrimPrefix(str, ".")
}

func GetFQN(fullName string) string {
	return StripStartingPeriod(fullName)
}

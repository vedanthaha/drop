package main

import (
	"fmt"
	"strings"
)

func ValidateFormatID(value string) bool {
	if value == "original" {
		return true
	}
	if value == "" || len(value) > 128 || strings.ContainsAny(value, "/\\\\\x00\r\n") {
		return false
	}
	for _, char := range value {
		if !(char == '+' || char == '-' || char == '_' || char == '.' || char >= '0' && char <= '9' || char >= 'A' && char <= 'Z' || char >= 'a' && char <= 'z') {
			return false
		}
	}
	return true
}

func main() {
	fmt.Println(ValidateFormatID("dash-1430991455627102v"))
}

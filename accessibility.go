package main

import (
	"os"
	"strconv"
)

func IsAccessible() bool {
	accessible, err := strconv.ParseBool(os.Getenv("ACCESSIBLE"))
	if err != nil {
		return false
	}
	return accessible
}

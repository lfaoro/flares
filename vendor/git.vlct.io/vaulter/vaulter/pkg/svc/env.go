package svc

import (
	"log"
	"os"
)

// MustGetEnv retrieves the provided environment variable
// value, and panics if not found.
func MustGetEnv(variable string) string {
	v := os.Getenv(variable)
	if v == "" {
		log.Panicf("%v environment variable not set.", variable)
	}
	return v
}

// MustGetEnvs retrieves the provided environment variables
// values, and panics if any is not found.
func MustGetEnvs(variables ...string) []string {
	var values []string
	for _, v := range variables {
		val := os.Getenv(v)
		if v == "" {
			log.Panicf("%v environment variable not set.", val)
		}
		values = append(values, val)
	}
	return values
}

package utils

import (
	"log"
)

func Contains(s *[]uint32, e uint32) bool {
	for _, a := range *s {
		if a == e {
			return true
		}
	}
	return false
}
func ErrorLog(err error, message string) {
	if err != nil {
		log.Fatal(message, err)
	}
}

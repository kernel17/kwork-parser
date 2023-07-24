package utils

import (
	"log"
	"strconv"
	"strings"
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

func UintSliceToString(slice []uint32) string {
	var IDs []string
	for _, i := range slice {
		IDs = append(IDs, strconv.Itoa(int(i)))
	}
	idsStr := strings.Join(IDs, ", ")
	return idsStr

}

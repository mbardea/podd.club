package logger

import (
	"log"
)

func Infof(format string, a ...interface{}) {
	log.Printf("INFO: "+format, a)
}

func Errorf(format string, a ...interface{}) {
	log.Printf("ERROR: "+format, a)
}

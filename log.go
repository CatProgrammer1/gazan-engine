package main

import "log"

func warn(msg ...any) {
	log.Println(msg...)
}

func Throw(msg ...any) {
	log.Fatalln(msg...)
}

func throwf(format string, msg ...any) {
	log.Fatalf(format, msg...)
}

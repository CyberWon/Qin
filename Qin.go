package main

import (
	"Qin/pkg/gateway"
	"log"
)

func main() {
	err := gateway.GS.Start()
	log.Fatal(err)
}

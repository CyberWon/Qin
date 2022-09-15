package main

import (
	"MicroOps/pkg/gateway"
	"log"
)

func main() {
	err := gateway.GS.Start()
	log.Fatal(err)
}

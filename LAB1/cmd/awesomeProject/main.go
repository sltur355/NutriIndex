package main

import (
	"LAB1/internal/api"
	"log"
)

func main() {
	log.Println("Appliation start!")
	api.StartServer()
	log.Println("Application terminated!")
}

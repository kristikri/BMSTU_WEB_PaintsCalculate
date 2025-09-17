package main

import (
	"log"
	"ssr_immemory/internal/api"
)
func main() {
	log.Println("Application start!")
	api.StartServer()
	log.Println("Application terminated!")
}

package main

import (
	"log"
	"net/http"

	"github.com/joshakeman/gametime-take-home/api"
	"github.com/joshakeman/gametime-take-home/data"
	"github.com/joshakeman/gametime-take-home/service"
)

func main() {
	store := data.NewStore()
	svc := service.New(store, service.FakePaymentService{}, service.FakeCompleter{})

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", api.NewHandler(svc)))
}

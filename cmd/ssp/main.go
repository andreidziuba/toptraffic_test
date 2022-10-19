package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/andreidziuba/toptraffic_test/cmd/ssp/components"
	flags "github.com/andreidziuba/toptraffic_test/internal/flags"
)

func main() {
	port, advertisingPartners := flags.FlagsParse()

	http.HandleFunc("/placements/request", components.NewHandleFunc(advertisingPartners))
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		log.Fatalf("error starting server: %s\n", err)
	}
}

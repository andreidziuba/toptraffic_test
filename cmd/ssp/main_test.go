package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andreidziuba/toptraffic_test/cmd/ssp/components"
	. "github.com/andreidziuba/toptraffic_test/pkg/structures"
)

func TestHandler(t *testing.T) {
	r := PlacementsRequest{
		Id: "780",
		Tiles: []Tiles{
			{
				Id:    15,
				Width: 100,
				Ratio: 1.5,
			},
			{
				Id:    17,
				Width: 110,
				Ratio: 1.2,
			}},
		Context: Context{
			Ip:        "192.168.10.10",
			UserAgent: "diospiros",
		},
	}
	s, _ := json.Marshal(r)
	req, err := http.NewRequest("GET", "/placements/request", bytes.NewReader(s))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	ap := make([]IPORT, 0)
	ap = append(ap, IPORT{
		IP:   []byte{127, 0, 0, 1},
		Port: 8000,
	})
	ap = append(ap, IPORT{
		IP:   []byte{127, 0, 0, 1},
		Port: 8001,
	})
	handler := http.HandlerFunc(components.NewHandleFunc(&ap))

	handler.ServeHTTP(rr, req)
	fmt.Println(rr.Body)
}

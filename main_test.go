package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	r := placements_request{
		Id: "780",
		Tiles: []tiles{tiles{
			Id:    15,
			Width: 100,
			Ratio: 1.5,
		}},
		Context: context{
			Ip:         "192.168.10.10",
			User_agent: "diospiros",
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
		port: 8000,
	})
	ap = append(ap, IPORT{
		IP:   []byte{127, 0, 0, 1},
		port: 8001,
	})
	handler := http.HandlerFunc(NewHandleFunc(ap))

	handler.ServeHTTP(rr, req)
	fmt.Println(rr.Body)
}

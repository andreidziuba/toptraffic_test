package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckHandler(t *testing.T) {
	// Создаем запрос с указанием нашего хендлера. Нам не нужно
	// указывать параметры, поэтому вторым аргументом передаем nil
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

	// Мы создаем ResponseRecorder(реализует интерфейс http.ResponseWriter)
	// и используем его для получения ответа
	rr := httptest.NewRecorder()
	ap := make([]IPORT, 0)
	ap = append(ap, IPORT{
		IP:   []byte{192, 168, 9, 9},
		port: 1900,
	})
	handler := http.HandlerFunc(NewHandleFunc(ap))

	// Наш хендлер соответствует интерфейсу http.Handler, а значит
	// мы можем использовать ServeHTTP и напрямую указать
	// Request и ResponseRecorder
	handler.ServeHTTP(rr, req)
	fmt.Println(rr.Body)
}

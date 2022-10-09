package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	port := flag.Int("p", -1, "port: 1-65535")
	advertising_partners_string := flag.String("d", "", "ip:port,ip:port,... [1-10]")
	advertising_partners := make([]IPORT, 0, 10)

	flag.Parse()
	if *port < 1 || *port > 65535 {
		fmt.Println("Port (flag -p) должен быть в пределах 1-65535")
		os.Exit(2)
	}
	//Парсим адреса рекламных партнёров
	for _, a := range strings.Split(*advertising_partners_string, ",") {
		a_split := strings.Split(a, ":")
		p, err := strconv.ParseUint(strings.Trim(a_split[1], " "), 10, 64)
		if err != nil || p < 1 || p > 65535 {
			fmt.Println("port не явлется числом от 1 до 65535: ", a)
			os.Exit(3)
		}
		ap_port := uint16(p)
		parse_ip := net.ParseIP(strings.Trim(a_split[0], " "))
		if parse_ip == nil {
			fmt.Println("Неверный ip", a_split[0])
			os.Exit(3)
		}
		advertising_partners = append(advertising_partners, IPORT{parse_ip, ap_port})
		if len(advertising_partners) > 10 {
			fmt.Println("Рекламных партнёров больше 10")
		}
		if len(advertising_partners) < 1 {
			fmt.Println("Рекламных партнёров меньше 1")
			os.Exit(3)
		}
	}
	http.HandleFunc("/placements/request", NewHandleFunc(advertising_partners))
	err := http.ListenAndServe(fmt.Sprintf("localhost:%d", *port), nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func NewHandleFunc(ap []IPORT) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		dec := json.NewDecoder(req.Body)
		dec.DisallowUnknownFields()
		json_request := placements_request{}
		err := dec.Decode(&json_request)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		if json_request.Id == "" {
			http.Error(rw, "(WRONG_SCHEMA) Нет поля 'Id' в JSON", http.StatusBadRequest)
			return
		}
		if json_request.Context == (context{}) {
			http.Error(rw, "(WRONG_SCHEMA) Нет поля 'Context' в JSON", http.StatusBadRequest)
			return
		}

		// TODO надо сделать также проверку полей в Context и Tiles
		// сделать свои поля по умолчанию для анмаршалинга и потом проверять на дефолтные значения.

		if len(json_request.Tiles) == 0 {
			http.Error(rw, "(EMPTY_TILES) Отстуствуют tiles", http.StatusBadRequest)
			return
		}

		// if dec.More() {
		// 	http.Error(rw, "Лишняя информация после JSON", http.StatusBadRequest)
		// 	return
		// }

		request_adverising_partners(rw, ap, json_request)
	}
}

func request_adverising_partners(rw http.ResponseWriter, ap []IPORT, pr placements_request) {
	breq := bid_request{Id: pr.Id, Context: pr.Context}

	for _, tiles := range pr.Tiles {
		ir := imp_request{
			Id:        tiles.Id,
			Minwidth:  tiles.Width,
			Minheight: uint(math.Floor(float64(tiles.Width) * tiles.Ratio)),
		}
		breq.Imp = append(breq.Imp, ir)
	}
	resp_chan := make(chan bid_response, 20)
	tr := &http.Transport{}
	client := &http.Client{Transport: tr, Timeout: 200 * time.Millisecond}
	var ap_wg sync.WaitGroup
	for _, iport := range ap {
		ap_wg.Add(1)
		go func(iport IPORT) {
			defer ap_wg.Done()
			b, err := json.Marshal(breq)
			if err != nil {
				fmt.Println("Error marshal:", err)
				panic(4)
			}
			r := bytes.NewReader(b)
			resp, err := client.Post(iport.to_url(), "application/json", r)
			if err != nil {
				fmt.Println(err)
				return
			}
			switch resp.StatusCode {
			case 200:
				dec := json.NewDecoder(resp.Body)
				dec.DisallowUnknownFields()
				json_request := bid_response{}
				err := dec.Decode(&json_request)
				if err != nil {
					fmt.Println("Error decode:", err)
					panic(4)
				}
				resp_chan <- json_request
			default:
				fmt.Println(resp.StatusCode, "фигня какая-то")
			}
		}(iport)
	}
	ap_wg.Wait()
	close(resp_chan)
	imp_bid_responses := make(map[uint]imp_bid_response)
	for bresp := range resp_chan {
		for _, imp := range bresp.Imp {
			if imp_bid_responses[imp.Id].Price < imp.Price {
				imp_bid_responses[imp.Id] = imp
			}
		}
	}
	pl_re := placements_response{
		Id: breq.Id,
	}
	for _, a := range pr.Tiles {
		temp_imp, ok := imp_bid_responses[a.Id]
		if !ok {
			continue
		}
		imp_resp := imp_response{
			Id:     temp_imp.Id,
			Width:  temp_imp.Width,
			Height: temp_imp.Height,
			Title:  temp_imp.Title,
			Url:    temp_imp.Url,
		}
		pl_re.Imp = append(pl_re.Imp, imp_resp)
	}
	json_pl_re, err := json.Marshal(pl_re)
	if err != nil {
		fmt.Println("Error marshal:", err)

	}
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(json_pl_re)
}

//	{
//		"id": <string>,
//		"tiles": [{
//			"id": <uint>,
//			"width": <uint>,
//			"ratio": <float>
//			}, … ],
//		"context": {
//			"ip": <ip4 string>,
//			"user_agent": <string>
//		}
//	}
type placements_request struct {
	Id      string  `json:"id"`
	Tiles   []tiles `json:"tiles"`
	Context context `json:"context"`
}

type tiles struct {
	Id    uint    `json:"id"`
	Width uint    `json:"width"`
	Ratio float64 `json:"ratio"`
}

type context struct {
	Ip         string `json:"ip"`
	User_agent string `json:"user_agent"`
}

// type string_ip struct {
// 	Ip string
// }

// POST /bid_request
// …
// Content-Type: application/json

//	{
//		"id": <string>,
//		"imp": [{
//			"id": <uint>,
//			"minwidth": <uint>,
//			"minheight": <uint>
//		}, … ],
//		"context": {
//			"ip": <ip4 string>,
//			"user_agent": <string>
//		}
//	}
type bid_request struct {
	Id      string        `json:"id"`
	Imp     []imp_request `json:"imp"`
	Context context       `json:"context"`
}

type imp_request struct {
	Id        uint `json:"id"`
	Minwidth  uint `json:"minwidth"`
	Minheight uint `json:"minheight"`
}

//	{
//		"id": <string>,
//		"imp": [{
//			"id": <uint>,
//			"width": <uint>,
//			"height": <uint>,
//			"title": <string>,
//			"url": <string>,
//			"price": <float>
//			}, … ]
//	}
type bid_response struct {
	Id  string             `json:"id"`
	Imp []imp_bid_response `json:"imp"`
}

type imp_bid_response struct {
	Id     uint    `json:"id"`
	Width  uint    `json:"width"`
	Height uint    `json:"height"`
	Title  string  `json:"title"`
	Url    string  `json:"url"`
	Price  float64 `json:"price"`
}

//	{
//		"id": <string>,
//		"imp": [{
//			"id": <uint>,
//			"width": <uint>,
//			"height": <uint>,
//			"title": <string>,
//			"url": <string>
//			}, … ]
//	}
type placements_response struct {
	Id  string         `json:"id"`
	Imp []imp_response `json:"imp"`
}

type imp_response struct {
	Id     uint   `json:"id"`
	Width  uint   `json:"width"`
	Height uint   `json:"height"`
	Title  string `json:"title"`
	Url    string `json:"url"`
}

type IPORT struct {
	IP   net.IP
	port uint16
}

func (iport *IPORT) to_string() string {
	return fmt.Sprintf("%s:%s", iport.IP.String(), strconv.FormatUint(uint64(iport.port), 10))
}

func (iport *IPORT) to_url() string {
	return fmt.Sprintf("http://%s", iport.to_string())
}

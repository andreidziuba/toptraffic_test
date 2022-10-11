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
	for _, ap_string := range strings.Split(*advertising_partners_string, ",") {
		ap_split_ip_port := strings.Split(ap_string, ":")
		p, err := strconv.ParseUint(strings.Trim(ap_split_ip_port[1], " "), 10, 64)
		if err != nil || p < 1 || p > 65535 {
			fmt.Println("port не явлется числом от 1 до 65535: ", ap_string)
			os.Exit(3)
		}
		ap_port := uint16(p)
		ap_parsed_ip := net.ParseIP(strings.Trim(ap_split_ip_port[0], " "))
		if ap_parsed_ip == nil {
			fmt.Println("Неверный ip", ap_split_ip_port[0])
			os.Exit(3)
		}
		advertising_partners = append(advertising_partners, IPORT{ap_parsed_ip, ap_port})
		if len(advertising_partners) > 10 {
			fmt.Println("Рекламных партнёров больше 10")
		}
		if len(advertising_partners) < 1 {
			fmt.Println("Рекламных партнёров меньше 1")
			os.Exit(3)
		}
	}
	http.HandleFunc("/placements/request", NewHandleFunc(&advertising_partners))
	err := http.ListenAndServe(fmt.Sprintf("localhost:%d", *port), nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func NewHandleFunc(ap *[]IPORT) func(http.ResponseWriter, *http.Request) {
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
		if json_request.Context.Ip == "" {
			http.Error(rw, "(EMPTY_FIELD) Нет поля 'ip' в Context", http.StatusBadRequest)
			return
		}
		if json_request.Context.User_agent == "" {
			http.Error(rw, "(EMPTY_FIELD) Нет поля 'User_agent' в Context", http.StatusBadRequest)
			return
		}

		// TODO надо сделать проверку полей в Context и Tiles
		// сделать свои поля по умолчанию для анмаршалинга и потом проверять на дефолтные значения.
		for _, tile := range json_request.Tiles {
			if tile.Id == 0 {
				http.Error(rw, "(EMPTY_FIELD) Нет поля 'Id' в Tile", http.StatusBadRequest)
				return
			}
			if tile.Ratio == 0 {
				http.Error(rw, "(EMPTY_FIELD) Нет поля 'Ratio' в Tile", http.StatusBadRequest)
				return
			}
			if tile.Width == 0 {
				http.Error(rw, "(EMPTY_FIELD) Нет поля 'Ratio' в Tile", http.StatusBadRequest)
				return
			}
		}

		if len(json_request.Tiles) == 0 {
			http.Error(rw, "(EMPTY_TILES) Отстуствуют tiles", http.StatusBadRequest)
			return
		}

		// if dec.More() {
		// 	http.Error(rw, "Лишняя информация после JSON", http.StatusBadRequest)
		// 	return
		// }

		request_adverising_partners(rw, ap, &json_request)
	}
}

func request_adverising_partners(rw http.ResponseWriter, ap *[]IPORT, pr *placements_request) {
	bid_req := bid_request{Id: pr.Id, Context: pr.Context}

	for _, tiles := range pr.Tiles {
		ir := imp_request{
			Id:        tiles.Id,
			Minwidth:  tiles.Width,
			Minheight: uint(math.Floor(float64(tiles.Width) * tiles.Ratio)),
		}
		bid_req.Imp = append(bid_req.Imp, ir)
	}
	resp_chan := make(chan bid_response, 20)
	client := &http.Client{Timeout: 200 * time.Millisecond}
	var ap_wg sync.WaitGroup
	for _, iport := range *ap {
		ap_wg.Add(1)
		go func(iport IPORT) {
			defer ap_wg.Done()
			b, err := json.Marshal(bid_req)
			if err != nil {
				fmt.Println("Error marshal:", err)
				panic(4)
			}
			r := bytes.NewReader(b)
			resp, err := client.Post(iport.to_url("bid_request"), "application/json", r)
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
		Id: bid_req.Id,
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
	Price  float64 `json:"price,string"`
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

func (iport *IPORT) to_url(a string) string {
	if a != "" {
		return fmt.Sprintf("http://%s/%s", iport.to_string(), a)
	}
	return fmt.Sprintf("http://%s", iport.to_string())
}

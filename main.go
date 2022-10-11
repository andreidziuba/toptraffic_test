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
	port, advertisingPartners := flagsParse()
	http.HandleFunc("/placements/request", NewHandleFunc(advertisingPartners))
	err := http.ListenAndServe(fmt.Sprintf("localhost:%d", *port), nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func flagsParse() (*int, *[]IPORT) {
	port := flag.Int("p", -1, "port: 1-65535")
	advertisingPartnersString := flag.String("d", "", "ip:port,ip:port,... [1-10]")
	advertisingPartners := make([]IPORT, 0, 10)

	flag.Parse()
	if *port < 1 || *port > 65535 {
		fmt.Println("Port (flag -p) должен быть в пределах 1-65535")
		os.Exit(2)
	}
	//Парсим адреса рекламных партнёров
	for _, apString := range strings.Split(*advertisingPartnersString, ",") {
		apSplitIpPort := strings.Split(apString, ":")
		p, err := strconv.ParseUint(strings.Trim(apSplitIpPort[1], " "), 10, 64)
		if err != nil || p < 1 || p > 65535 {
			fmt.Println("port не явлется числом от 1 до 65535: ", apString)
			os.Exit(3)
		}
		apPort := uint16(p)
		apParsedIp := net.ParseIP(strings.Trim(apSplitIpPort[0], " "))
		if apParsedIp == nil {
			fmt.Println("Неверный ip", apSplitIpPort[0])
			os.Exit(3)
		}
		advertisingPartners = append(advertisingPartners, IPORT{apParsedIp, apPort})
		if len(advertisingPartners) > 10 {
			fmt.Println("Рекламных партнёров больше 10")
		}
		if len(advertisingPartners) < 1 {
			fmt.Println("Рекламных партнёров меньше 1")
			os.Exit(3)
		}
	}
	return port, &advertisingPartners
}

func NewHandleFunc(ap *[]IPORT) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		dec := json.NewDecoder(req.Body)
		dec.DisallowUnknownFields()
		jsonRequest := placementsRequest{}
		err := dec.Decode(&jsonRequest)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		if jsonRequest.Id == "" {
			http.Error(rw, "(WRONG_SCHEMA) Нет поля 'Id' в JSON", http.StatusBadRequest)
			return
		}
		if jsonRequest.Context == (context{}) {
			http.Error(rw, "(WRONG_SCHEMA) Нет поля 'Context' в JSON", http.StatusBadRequest)
			return
		}
		if jsonRequest.Context.Ip == "" {
			http.Error(rw, "(EMPTY_FIELD) Нет поля 'ip' в Context", http.StatusBadRequest)
			return
		}
		if jsonRequest.Context.UserAgent == "" {
			http.Error(rw, "(EMPTY_FIELD) Нет поля 'User_agent' в Context", http.StatusBadRequest)
			return
		}

		// TODO надо сделать проверку полей в Context и Tiles
		// сделать свои поля по умолчанию для анмаршалинга и потом проверять на дефолтные значения.
		for _, tile := range jsonRequest.Tiles {
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

		if len(jsonRequest.Tiles) == 0 {
			http.Error(rw, "(EMPTY_TILES) Отстуствуют tiles", http.StatusBadRequest)
			return
		}

		// if dec.More() {
		// 	http.Error(rw, "Лишняя информация после JSON", http.StatusBadRequest)
		// 	return
		// }

		requestAdvertisingPartners(rw, ap, &jsonRequest)
	}
}

func requestAdvertisingPartners(rw http.ResponseWriter, advertisingPartners *[]IPORT, pr *placementsRequest) {
	bidReq := bidRequest{Id: pr.Id, Context: pr.Context}

	for _, tiles := range pr.Tiles {
		ir := impRequest{
			Id:        tiles.Id,
			Minwidth:  tiles.Width,
			Minheight: uint(math.Floor(float64(tiles.Width) * tiles.Ratio)),
		}
		bidReq.Imp = append(bidReq.Imp, ir)
	}
	respChan := make(chan bidResponse, 20)
	client := &http.Client{Timeout: 200 * time.Millisecond}
	var apWG sync.WaitGroup
	for _, apIPORT := range *advertisingPartners {
		apWG.Add(1)
		go func(iport IPORT) {
			defer apWG.Done()
			b, err := json.Marshal(bidReq)
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
				jsonRequest := bidResponse{}
				err := dec.Decode(&jsonRequest)
				if err != nil {
					fmt.Println("Error decode:", err)
					panic(4)
				}
				respChan <- jsonRequest
			default:
				fmt.Println(resp.StatusCode, "фигня какая-то")
			}
		}(apIPORT)
	}
	apWG.Wait()
	close(respChan)
	impBidResponses := make(map[uint]impBidResponse)
	for bidResp := range respChan {
		for _, imp := range bidResp.Imp {
			if impBidResponses[imp.Id].Price < imp.Price {
				impBidResponses[imp.Id] = imp
			}
		}
	}
	plRe := placementsResponse{
		Id: bidReq.Id,
	}
	for _, a := range pr.Tiles {
		tempImp, ok := impBidResponses[a.Id]
		if !ok {
			continue
		}
		impResp := impResponse{
			Id:     tempImp.Id,
			Width:  tempImp.Width,
			Height: tempImp.Height,
			Title:  tempImp.Title,
			Url:    tempImp.Url,
		}
		plRe.Imp = append(plRe.Imp, impResp)
	}
	jsonPlRe, err := json.Marshal(plRe)
	if err != nil {
		fmt.Println("Error marshal:", err)

	}
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(jsonPlRe)
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
type placementsRequest struct {
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
	Ip        string `json:"ip"`
	UserAgent string `json:"user_agent"`
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
type bidRequest struct {
	Id      string       `json:"id"`
	Imp     []impRequest `json:"imp"`
	Context context      `json:"context"`
}

type impRequest struct {
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
type bidResponse struct {
	Id  string           `json:"id"`
	Imp []impBidResponse `json:"imp"`
}

type impBidResponse struct {
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
type placementsResponse struct {
	Id  string        `json:"id"`
	Imp []impResponse `json:"imp"`
}

type impResponse struct {
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

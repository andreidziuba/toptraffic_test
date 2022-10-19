package structures

import (
	"fmt"
	"net"
	"strconv"
)

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
type PlacementsRequest struct {
	Id      string  `json:"id"`
	Tiles   []Tiles `json:"tiles"`
	Context Context `json:"context"`
}

type Tiles struct {
	Id    uint    `json:"id"`
	Width uint    `json:"width"`
	Ratio float64 `json:"ratio"`
}

type Context struct {
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
type BidRequest struct {
	Id      string       `json:"id"`
	Imp     []ImpRequest `json:"imp"`
	Context Context      `json:"context"`
}

type ImpRequest struct {
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
type BidResponse struct {
	Id  string           `json:"id"`
	Imp []ImpBidResponse `json:"imp"`
}

type ImpBidResponse struct {
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
type PlacementsResponse struct {
	Id  string        `json:"id"`
	Imp []ImpResponse `json:"imp"`
}

type ImpResponse struct {
	Id     uint   `json:"id"`
	Width  uint   `json:"width"`
	Height uint   `json:"height"`
	Title  string `json:"title"`
	Url    string `json:"url"`
}

type IPORT struct {
	IP   net.IP
	Port uint16
}

func (iport *IPORT) to_string() string {
	return fmt.Sprintf("%s:%s", iport.IP.String(), strconv.FormatUint(uint64(iport.Port), 10))
}

func (iport *IPORT) To_url(a string) string {
	if a != "" {
		return fmt.Sprintf("http://%s/%s", iport.to_string(), a)
	}
	return fmt.Sprintf("http://%s", iport.to_string())
}

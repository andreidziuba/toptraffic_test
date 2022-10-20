package structures

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

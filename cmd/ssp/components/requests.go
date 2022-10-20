package components

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"net/http"
	"sync"
	"time"

	. "github.com/andreidziuba/toptraffic_test/pkg/structures"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func requestAdvertisingPartners(rw http.ResponseWriter, advertisingPartners *[]IPORT, pr *PlacementsRequest) (*chan BidResponse, *BidRequest) {
	bidReq := prepareBidRequest(pr)

	respChan := make(chan BidResponse, 20)
	client := &http.Client{Timeout: 200 * time.Millisecond}
	var apWG sync.WaitGroup
	for _, apIPORT := range *advertisingPartners {
		apWG.Add(1)
		go requestAdvertisingPartner(&apWG, bidReq, client, respChan, apIPORT)
	}
	apWG.Wait()
	close(respChan)

	return &respChan, &bidReq
}

func prepareBidRequest(pr *PlacementsRequest) BidRequest {
	bidReq := BidRequest{Id: pr.Id, Context: pr.Context}

	for _, tiles := range pr.Tiles {
		ir := ImpBidRequest{
			Id:        tiles.Id,
			Minwidth:  tiles.Width,
			Minheight: uint(math.Floor(float64(tiles.Width) * tiles.Ratio)),
		}
		bidReq.Imp = append(bidReq.Imp, ir)
	}
	return bidReq
}

// TODO shouldbereturn
func requestAdvertisingPartner(apWG *sync.WaitGroup, bidReq BidRequest, client *http.Client, respChan chan BidResponse, iport IPORT) {
	defer apWG.Done()
	b, err := json.Marshal(bidReq)
	if err != nil {
		log.Println("Error marshal:", err)
		panic(4)
	}
	r := bytes.NewReader(b)
	resp, err := client.Post(iport.To_url("bid_request"), "application/json", r)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		dec := json.NewDecoder(resp.Body)
		dec.DisallowUnknownFields()
		jsonRequest := BidResponse{}
		err := dec.Decode(&jsonRequest)
		if err != nil {
			log.Println("Error decode:", err)
			panic(4)
		}
		respChan <- jsonRequest
	case 204:
		break
	default:
		fmt.Println(resp.StatusCode, "фигня какая-то")
	}
}

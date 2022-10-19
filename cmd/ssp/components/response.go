package components

import (
	"log"
	"net/http"

	. "github.com/andreidziuba/toptraffic_test/pkg/structures"
)

func placementResponse(plRe *PlacementsResponse, rw http.ResponseWriter) {
	jsonPlRe, err := json.Marshal(plRe)
	if err != nil {
		log.Println("Error marshal:", err)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(jsonPlRe)
}

func prepareBidResponse(respChan *chan BidResponse, bidReq *BidRequest, pr *PlacementsRequest) *PlacementsResponse {
	impBidResponses := make(map[uint]ImpBidResponse)
	for bidResp := range *respChan {
		for _, imp := range bidResp.Imp {
			if impBidResponses[imp.Id].Price < imp.Price {
				impBidResponses[imp.Id] = imp
			}
		}
	}
	plRe := PlacementsResponse{
		Id: bidReq.Id,
	}
	for _, a := range pr.Tiles {
		tempImp, ok := impBidResponses[a.Id]
		if !ok {
			continue
		}
		impResp := ImpResponse{
			Id:     tempImp.Id,
			Width:  tempImp.Width,
			Height: tempImp.Height,
			Title:  tempImp.Title,
			Url:    tempImp.Url,
		}
		plRe.Imp = append(plRe.Imp, impResp)
	}
	return &plRe
}

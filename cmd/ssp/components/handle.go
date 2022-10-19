package components

import (
	"log"
	"net/http"

	. "github.com/andreidziuba/toptraffic_test/pkg/structures"
)

func NewHandleFunc(ap *[]IPORT) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		jsonRequest, err := unmarshalAndCheckPlacementRequest(req, rw)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			log.Println(err.Error())
			return
		}
		// опрашиваем рекламных партнёров
		respChan, bidReq := requestAdvertisingPartners(rw, ap, jsonRequest)
		// подготавливаем ответ рекламным площадкам
		plRe := prepareBidResponse(respChan, bidReq, jsonRequest)
		// отвечаем рекламным площадкам
		placementResponse(plRe, rw)
	}
}

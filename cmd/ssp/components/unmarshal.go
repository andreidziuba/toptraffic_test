package components

import (
	"errors"
	"net/http"

	. "github.com/andreidziuba/toptraffic_test/pkg/structures"
)

func unmarshalAndCheckPlacementRequest(req *http.Request, rw http.ResponseWriter) (*PlacementsRequest, error) {
	jsonRequest, err := unmarshalPlacementRequest(req, rw)
	if err != nil {
		return &PlacementsRequest{}, errors.New("WRONG_SCHEMA")
	}
	err = checkJson(jsonRequest, rw)
	if err != nil {
		return &PlacementsRequest{}, err
	}
	return jsonRequest, nil
}

func unmarshalPlacementRequest(req *http.Request, rw http.ResponseWriter) (*PlacementsRequest, error) {
	dec := json.NewDecoder(req.Body)
	defer req.Body.Close()
	dec.DisallowUnknownFields()
	jsonRequest := PlacementsRequest{
		Id:    "**not_exist**",
		Tiles: []Tiles{},
		Context: Context{
			Ip:        "**not_exist**",
			UserAgent: "**not_exist**",
		},
	}
	err := dec.Decode(&jsonRequest)
	if err != nil {
		return &PlacementsRequest{}, err
	}
	return &jsonRequest, nil
}

func checkJson(jsonRequest *PlacementsRequest, rw http.ResponseWriter) error {
	if jsonRequest.Id == "**not_exist**" {
		return errors.New("(WRONG_SCHEMA) Нет поля 'Id' в JSON")
	}
	err := checkContext(jsonRequest)
	if err != nil {
		return err
	}
	err = checkTiles(jsonRequest)
	if err != nil {
		return err
	}
	return nil
}

func checkContext(jsonRequest *PlacementsRequest) error {
	if jsonRequest.Context == (Context{
		Ip:        "**not_exist**",
		UserAgent: "**not_exist**",
	}) {
		return errors.New("(WRONG_SCHEMA) Нет поля 'Context' в JSON")
	}
	if jsonRequest.Context.Ip == "" {
		return errors.New("(EMPTY_FIELD) Нет поля 'ip' в Context")
	}
	if jsonRequest.Context.UserAgent == "" {
		return errors.New("(EMPTY_FIELD) Нет поля 'User_agent' в Context")
	}
	return nil
}

func checkTiles(jsonRequest *PlacementsRequest) error {
	for _, tile := range jsonRequest.Tiles {
		if tile.Id == 0 {
			return errors.New("(EMPTY_FIELD) Нет поля 'Id' в Tile")
		}
		if tile.Ratio == 0 {
			return errors.New("(EMPTY_FIELD) Нет поля 'Ratio' в Tile")
		}
		if tile.Width == 0 {
			return errors.New("(EMPTY_FIELD) Нет поля 'Width' в Tile")
		}
	}
	if len(jsonRequest.Tiles) == 0 {
		return errors.New("(EMPTY_TILES) Отстуствуют tiles")
	}
	return nil
}

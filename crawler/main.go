package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type SearchFormFilter struct {
	Identificador string `json:"identificador"`
}
type SearchForm struct {
	Filter       string `json:"filter"`
	TipoBusqueda string `json:"tipoBusqueda"`
	Page         int    `json:"page"`
	Size         int    `json:"size"`
}

type SearchResultContent struct {
	ExpId int `json:"expId"`
}

type SearchResult struct {
	Content []SearchResultContent `json:"content"`
}

func searchExpediente(criteria string) (interface{}, error) {
	filter, _ := json.Marshal(SearchFormFilter{
		Identificador: criteria,
	})
	info, _ := json.Marshal(SearchForm{
		Filter:       string(filter),
		TipoBusqueda: "CAU",
		Page:         0,
		Size:         10,
	})

	resp, err := http.PostForm("https://eje.juscaba.gob.ar/iol-api/api/public/expedientes/lista", url.Values{
		"info": {string(info)},
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	sr := SearchResult{}
	err = json.NewDecoder(resp.Body).Decode(&sr)
	if err != nil {
		return nil, err
	}
	return sr, nil
}

func main() {
	res, err := searchExpediente("182908/2020-0")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v \n", res)
}

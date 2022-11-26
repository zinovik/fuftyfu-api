// /api/index.go
package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Data struct {
	En string `json:"en"`
	Ru string `json:"ru"`
}

type Hedgehog struct {
	Id          int        `json:"id"`
	When        string     `json:"when"`
	Photo       string     `json:"photo"`
	Who         Data       `json:"who"`
	Country     Data       `json:"country"`
	Place       Data       `json:"place"`
	Comment     Data       `json:"comment"`
	Coordinates [2]float32 `json:"coordinates"`
}

type Response struct {
	Total     int        `json:"total"`
	Filtered  int        `json:"filtered"`
	Hedgehogs []Hedgehog `json:"hedgehogs"`
}

const JSON_URL = "https://raw.githubusercontent.com/zinovik/fuftyfu-data/main/hedgehogs.json"

func getAllHedgehogs() []Hedgehog {
	response, err := http.Get(JSON_URL)
	if err != nil {
		fmt.Println(err)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}

	var hedgehogs []Hedgehog
	json.Unmarshal(body, &hedgehogs)

	return hedgehogs
}

func getResponse(hedgehogs []Hedgehog, limit, offset int, filter string) Response {
	var hedgehogsFiltered []Hedgehog

	for _, hedgehog := range hedgehogs {
		if strings.Contains(strings.ToLower(hedgehog.Country.En), filter) ||
			strings.Contains(strings.ToLower(hedgehog.Country.Ru), filter) ||
			strings.Contains(strings.ToLower(hedgehog.Place.En), filter) ||
			strings.Contains(strings.ToLower(hedgehog.Place.Ru), filter) ||
			strings.Contains(strings.ToLower(hedgehog.Comment.En), filter) ||
			strings.Contains(strings.ToLower(hedgehog.Comment.Ru), filter) ||
			strings.Contains(strings.ToLower(hedgehog.Who.En), filter) ||
			strings.Contains(strings.ToLower(hedgehog.Who.Ru), filter) {
			hedgehogsFiltered = append(hedgehogsFiltered, hedgehog)
		}
	}

	var response Response

	response.Total = len(hedgehogs)
	response.Filtered = len(hedgehogsFiltered)
	if offset >= len(hedgehogsFiltered) {
		response.Hedgehogs = []Hedgehog{}
	} else {
		response.Hedgehogs = hedgehogsFiltered[offset:int(math.Min(float64(offset+limit), float64(len(hedgehogsFiltered))))]
	}

	return response
}

func Main(w http.ResponseWriter, r *http.Request) {
	var token = r.URL.Query().Get("token")
	var isCors = r.URL.Query().Get("cors") == "true"

	w.Header().Set("Content-Type", "application/json")
	if isCors {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}

	if token != os.Getenv("API_TOKEN") {
		w.WriteHeader(401)
		fmt.Fprintf(w, "\"wrong token\"")
		return
	}

	var hedgehogs = getAllHedgehogs()

	if strings.HasPrefix(r.URL.Path, "/api/hedgehog") {
		var idString = strings.Replace(r.URL.Path, "/api/hedgehog/", "", 1)

		id, err := strconv.Atoi(idString)
		if err != nil || len(hedgehogs) < id {
			fmt.Println(id)
			w.WriteHeader(404)
			fmt.Fprint(w, "\"not found\"")
			return
		}

		var response, stringifyError = json.Marshal(hedgehogs[id-1])
		if stringifyError != nil {
			fmt.Println(stringifyError)
		}

		fmt.Fprint(w, string(response))
		return
	}

	var limitString = r.URL.Query().Get("limit")
	var offsetString = r.URL.Query().Get("offset")
	var filter = strings.ToLower(r.URL.Query().Get("filter"))

	limit, err := strconv.Atoi(limitString)
	if err != nil {
		limit = 10
	}
	offset, err := strconv.Atoi(offsetString)
	if err != nil {
		offset = 0
	}
	var responseObj = getResponse(hedgehogs, limit, offset, filter)

	var response, stringifyError = json.Marshal(responseObj)
	if stringifyError != nil {
		fmt.Println(stringifyError)
	}

	fmt.Fprint(w, string(response))
}

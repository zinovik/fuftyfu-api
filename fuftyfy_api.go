package fuftyfy_api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/storage"
)

type Data struct {
	En string `json:"en"`
	Ru string `json:"ru"`
}

type Hedgehog struct {
	Id          int        `json:"id"`
	When        string     `json:"when"`
	Photos      [2]string  `json:"photos"`
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

const BUCKET = "hedgehogs"
const FILE = "hedgehogs.json"

var (
	storageClient *storage.Client
	clientOnce    sync.Once
	clientErr     error
)

func getStorageClient() (*storage.Client, error) {
	clientOnce.Do(func() {
		storageClient, clientErr = storage.NewClient(context.Background())
	})
	return storageClient, clientErr
}

func parseGCSURL(rawURL string) (string, string, bool) {
	if strings.HasPrefix(rawURL, "https://storage.googleapis.com/") {
		trimmed := strings.TrimPrefix(rawURL, "https://storage.googleapis.com/")
		parts := strings.SplitN(trimmed, "/", 2)
		if len(parts) == 2 {
			return parts[0], parts[1], true
		}
	} else if strings.HasPrefix(rawURL, "gs://") {
		trimmed := strings.TrimPrefix(rawURL, "gs://")
		parts := strings.SplitN(trimmed, "/", 2)
		if len(parts) == 2 {
			return parts[0], parts[1], true
		}
	}
	return "", "", false
}

func signHedgehogPhotos(ctx context.Context, client *storage.Client, h *Hedgehog, saEmail string) {
	for i, photoURL := range h.Photos {
		if photoURL == "" {
			continue
		}
		bucket, object, ok := parseGCSURL(photoURL)
		if !ok {
			continue
		}
		opts := &storage.SignedURLOptions{
			Scheme:  storage.SigningSchemeV4,
			Method:  "GET",
			Expires: time.Now().Add(15 * time.Minute),
		}
		if saEmail != "" {
			opts.GoogleAccessID = saEmail
		}
		signedURL, err := client.Bucket(bucket).SignedURL(object, opts)
		if err != nil {
			log.Printf("Failed to generate signed URL for %s: %v", photoURL, err)
			continue
		}
		h.Photos[i] = signedURL
	}
}

func getAllHedgehogs(ctx context.Context, client *storage.Client) []Hedgehog {
	rc, err := client.Bucket(BUCKET).Object(FILE).NewReader(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer rc.Close()

	body, err := io.ReadAll(rc)
	if err != nil {
		log.Fatal(err)
	}

	var hedgehogs []Hedgehog
	json.Unmarshal(body, &hedgehogs)

	for i, j := 0, len(hedgehogs)-1; i < j; i, j = i+1, j-1 {
		hedgehogs[i], hedgehogs[j] = hedgehogs[j], hedgehogs[i]
	}

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

func init() {
	functions.HTTP("main", main)
}

func main(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Triggered!")

	var token = r.URL.Query().Get("token")
	var isCors = r.URL.Query().Get("cors") == "true"

	w.Header().Set("Content-Type", "application/json")
	if isCors {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}

	if token != os.Getenv("TOKEN") {
		w.WriteHeader(401)
		fmt.Fprintf(w, "\"wrong token\"")
		return
	}

	client, err := getStorageClient()
	if err != nil {
		log.Printf("Failed to get storage client: %v", err)
		w.WriteHeader(500)
		fmt.Fprint(w, "\"internal server error\"")
		return
	}

	var hedgehogs = getAllHedgehogs(r.Context(), client)

	// Fetch service account email from metadata (if running on GCF) for signing.
	saEmail, _ := metadata.Email("default")

	if strings.HasPrefix(r.URL.Path, "/api/hedgehog") {
		var idString = strings.Replace(r.URL.Path, "/api/hedgehog/", "", 1)

		id, err := strconv.Atoi(idString)
		if err != nil || len(hedgehogs) < id {
			fmt.Println(id)
			w.WriteHeader(404)
			fmt.Fprint(w, "\"not found\"")
			return
		}

		hedgehog := hedgehogs[id-1]
		signHedgehogPhotos(r.Context(), client, &hedgehog, saEmail)

		var response, stringifyError = json.Marshal(hedgehog)
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

	// Sign photos only for the hedgehogs being returned in the response page.
	for i := range responseObj.Hedgehogs {
		signHedgehogPhotos(r.Context(), client, &responseObj.Hedgehogs[i], saEmail)
	}

	var response, stringifyError = json.Marshal(responseObj)
	if stringifyError != nil {
		fmt.Println(stringifyError)
	}

	fmt.Fprint(w, string(response))

	fmt.Println("Done!")
}

package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// Struct auxiliar que será utilizada para armazenar apenas o Bid e codificar em json
type Bid struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, error := io.ReadAll(resp.Body)
	if error != nil {
		panic(error)
	}

	var ColetaBid Bid
	err = json.Unmarshal(body, &ColetaBid)
	if err != nil {
		panic(err)
	}

	println(ColetaBid)

}

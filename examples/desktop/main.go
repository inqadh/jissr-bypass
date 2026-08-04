package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	bypass "github.com/jissr-org/jissr-bypass"
)

func main() {
	request, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"https://blocked.yusef.ren/v1/health",
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Accept", "application/json")
	response, err := bypass.NewDefaultClient().Do(request)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()
	fmt.Printf("ECH request returned HTTP %d\n", response.StatusCode)
}

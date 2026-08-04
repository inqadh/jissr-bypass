package main

import (
	"fmt"
	"log"

	bypass "github.com/jissr-org/jissr-bypass"
)

func main() {
	result, err := bypass.NewDefaultClient().
		VerifyCanary("https://blocked.yusef.ren/v1/canary")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ECH canary verified: %s at %s\n", result.Host, result.IssuedAt)
}

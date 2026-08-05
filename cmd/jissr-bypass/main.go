package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	bypass "github.com/inqadh/jissr-bypass"
)

func main() {
	canary := flag.String("canary", "https://blocked.yusef.ren/v1/canary", "canary URL")
	host := flag.String("probe", "", "probe a hostname instead of verifying the canary")
	flag.Parse()

	client := bypass.NewDefaultClient()
	var value any
	var err error
	if *host != "" {
		value = client.Probe(*host)
		if !value.(*bypass.ProbeResult).OK {
			err = fmt.Errorf("%s: %s", value.(*bypass.ProbeResult).Stage, value.(*bypass.ProbeResult).Detail)
		}
	} else {
		value, err = client.VerifyCanary(*canary)
	}
	output, _ := json.MarshalIndent(value, "", "  ")
	fmt.Println(string(output))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

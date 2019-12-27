package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/r3boot/as65342-netbox/lib/generator"

	httptransport "github.com/go-openapi/runtime/client"

	"github.com/r3boot/as65342-netbox/lib/common"
	"github.com/r3boot/as65342-netbox/lib/netbox/client"
	"github.com/r3boot/as65342-netbox/lib/netboxclient"
)

const (
	netboxHostDefault  = "localhost:443"
	netboxTokenDefault = ""
	netboxNoTLSDefault = false
)

func main() {
	netboxHost := flag.String("api", netboxHostDefault, "Api host:port (NETBOX_HOST)")
	netboxToken := flag.String("token", netboxTokenDefault, "Token to use (NETBOX_TOKEN)")
	netboxNoTLS := flag.Bool("notls", netboxNoTLSDefault, "Set to disable TLS")
	netboxOutput := flag.String("out", "", "Where to store output")
	flag.Parse()

	http_proto := "https"
	if *netboxNoTLS {
		fmt.Printf("WARNING: disabling TLS!\n")
		http_proto = "http"
	}

	host := *netboxHost
	envHost := os.Getenv("NETBOX_HOST")
	if envHost != "" && *netboxHost == netboxHostDefault {
		host = envHost
	}

	token := *netboxToken
	envToken := os.Getenv("NETBOX_TOKEN")
	if envToken != "" && *netboxToken == netboxTokenDefault {
		token = envToken
	}

	transport := httptransport.New(host, client.DefaultBasePath, []string{http_proto})

	netbox, err := netboxclient.NewNetboxClient(
		client.New(transport, nil),
		common.NewTokenAuth(token),
		int64(9999),
	)
	if err != nil {
		fmt.Printf("ERROR: NewNetboxClient: %v\n", err)
		os.Exit(1)
	}

	generate, err := generator.NewGenerator(netbox, *netboxOutput)
	if err != nil {
		fmt.Printf("ERROR: NewGenerator: %v\n", err)
		os.Exit(1)
	}

	if err := generate.BackupHosts(); err != nil {
		fmt.Printf("ERROR: BackupHosts: %v\n", err)
		os.Exit(1)
	}
}

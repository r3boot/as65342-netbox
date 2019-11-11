package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/r3boot/as65342-netbox/lib/common"
	"github.com/r3boot/as65342-netbox/lib/generator"
	"github.com/r3boot/as65342-netbox/lib/netboxclient"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/r3boot/as65342-netbox/lib/netbox/client"
)

const (
	netboxHostDefault         = "localhost:443"
	netboxTokenDefault        = ""
	netboxNoTLSDefault        = false
	netboxOperationReverseDNS = "rdns"
	netboxOperationAnsible    = "ansible"
	netboxOperationIcinga2    = "icinga2"
)

func main() {
	netboxHost := flag.String("api", netboxHostDefault, "Api host:port (NETBOX_HOST)")
	netboxToken := flag.String("token", netboxTokenDefault, "Token to use (NETBOX_TOKEN)")
	netboxNoTLS := flag.Bool("notls", netboxNoTLSDefault, "Set to disable TLS")
	netboxGenerate := flag.String("generate", "", "What to generate")
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

	if *netboxOutput == "" {
		fmt.Printf("ERROR: Need an output path\n")
		os.Exit(1)
	}

	generate, err := generator.NewGenerator(netbox, *netboxOutput)
	if err != nil {
		fmt.Printf("ERROR: NewGenerator: %v\n", err)
		os.Exit(1)
	}

	switch *netboxGenerate {
	case netboxOperationAnsible:
		{
			if err := generate.AnsibleInventory(); err != nil {
				fmt.Printf("ERROR: AnsibleInventory: %v\n", err)
				os.Exit(1)
			}

			if err = generate.AnsibleGroupVars(); err != nil {
				fmt.Printf("ERROR: AnsibleGroupVars: %v\n", err)
				os.Exit(1)
			}

			if err = generate.AnsibleHostVars(); err != nil {
				fmt.Printf("ERROR: AnsibleHostVars: %v\n", err)
				os.Exit(1)
			}
		}
	case netboxOperationIcinga2:
		{
			if err := generate.Icinga2Config(); err != nil {
				fmt.Printf("ERROR: Icinga2Config: %v\n", err)
				os.Exit(1)
			}
		}
	case netboxOperationReverseDNS:
		{
			serial := time.Now().Format("20060102150405")
			/*
				if err := generate.ReverseDNS(serial); err != nil {
					fmt.Printf("ERROR: ReverseDNS: %v\n", err)
					os.Exit(1)
				}
			*/

			if err := generate.ForwardDNS(serial); err != nil {
				fmt.Printf("ERROR: ForwardDNS: %v\n", err)
				os.Exit(1)
			}
		}
	default:
		{
			fmt.Printf("Dont know how to generate %v\n", netboxGenerate)
			os.Exit(1)
		}
	}
}

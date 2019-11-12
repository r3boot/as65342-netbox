package generator

import (
	"fmt"
	"html/template"
	"os"

	"strings"

	"github.com/r3boot/as65342-netbox/lib/common"
)

const (
	PTR = "PTR"
)

const reverseDnsZoneTemplate = `$ORIGIN .
$TTL 60 ; 1 minute
{{ .Name }}   IN SOA  master.as65342.net. hostmaster.as65342.net. (
                                {{ .Serial }} ; serial
                                3600       ; refresh (1 hour)
                                7200       ; retry (2 hours)
                                2419200    ; expire (4 weeks)
                                60         ; minimum (1 minute)
                                )
                        NS      ns.as65342.net.
$ORIGIN {{ .Name }}.
* PTR unallocated.as65342.net.
{{- range .Records }}
{{ .Name }} {{ .Type }} {{ .Value }}
{{- end }}
`

const forwardDnsZoneTemplate = `$ORIGIN .
$TTL 60 ; 1 minute
{{ .Name }}   IN SOA  master.as65342.net. hostmaster.as65342.net. (
                                {{ .Serial }} ; serial
                                3600       ; refresh (1 hour)
                                7200       ; retry (2 hours)
                                2419200    ; expire (4 weeks)
                                60         ; minimum (1 minute)
                                )
                        NS      ns.as65342.net.
                        MX      10 mail.as65342.net.
$ORIGIN {{ .Name }}.
{{- range .Records }}
{{ .Name }} {{ .Type }} {{ .Value }}
{{- end }}
`

type Record struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Zone struct {
	Name    string   `json:"name"`
	Records []Record `json:"records"`
}

type ZonesConfig map[string]interface{}

type dnsZoneParams struct {
	Name    string
	Serial  string
	Records []Record
}

func (g Generator) ReverseDNS(serial string) error {
	allPrefixes, err := g.client.GetPrefixList("as65342")
	if err != nil {
		return fmt.Errorf("GetPrefixList: %v", err)
	}

	allIpAddresses, err := g.client.GetIpAddressList("as65342")
	if err != nil {
		return fmt.Errorf("GetIpAddressList: %v", err)
	}

	allZones := []Zone{}
	for _, prefix := range allPrefixes {
		zone := Zone{
			Name:    common.ToDnsZoneName(prefix),
			Records: []Record{},
		}

		for _, ipAddress := range allIpAddresses {
			if prefix.Contains(ipAddress.Address) {
				name := ""
				if strings.Contains(ipAddress.Address.String(), ":") {
					i := ipAddress.Address
					tmp := fmt.Sprintf("%02x%02x%02x%02x%02x%02x%02x%02x", i[8], i[9], i[10], i[11], i[12], i[13], i[14], i[15])
					for _, v := range tmp {
						if name == "" {
							name = string(v)
						} else {
							name = string(v) + "." + name
						}
					}
				} else {
					name = fmt.Sprintf("%v", ipAddress.Address.To4()[3])
				}
				r := Record{
					Name:  name,
					Type:  PTR,
					Value: common.ToFqdn(ipAddress.Dns),
				}
				zone.Records = append(zone.Records, r)
			}
		}

		allZones = append(allZones, zone)
	}

	err = common.CreateDirIfNotExists(g.out)
	if err != nil {
		return fmt.Errorf("CreateDirIfNotExists: %v\n", err)
	}

	for _, zone := range allZones {
		t, err := template.New("reverseDnsZoneTemplate").Parse(reverseDnsZoneTemplate)
		if err != nil {
			return fmt.Errorf("template.New: %v", err)
		}

		p := dnsZoneParams{
			Name:    zone.Name,
			Serial:  serial,
			Records: zone.Records,
		}

		fname := g.out + "/db." + zone.Name
		fd, err := os.Create(fname + ".new")
		if err != nil {
			return fmt.Errorf("os.Open: %v", err)
		}
		defer func() {
			fd.Close()
			os.Rename(fname+".new", fname)
			fmt.Printf("[+] Wrote %s\n", fname)
		}()

		err = t.Execute(fd, p)
		if err != nil {
			return fmt.Errorf("t.Execute: %v", err)
		}
	}

	return nil
}

func (g Generator) ForwardDNS(serial string) error {
	allConfigContexts, err := g.client.ListConfigContexts()
	if err != nil {
		return fmt.Errorf("ListConfigContexts: %v", err)
	}

	allIpAddresses, err := g.client.GetIpAddressList("as65342")
	if err != nil {
		return fmt.Errorf("GetIpAddressList: %v", err)
	}

	allZonesNoHosts := []Zone{}
	for _, context := range allConfigContexts {
		if context.Name == "dns_zones" {
			zones := context.Config.(map[string]interface{})["dns_zones"].([]interface{})
			for _, zoneData := range zones {
				zone := zoneData.(map[string]interface{})
				newZone := Zone{
					Name:    zone["name"].(string),
					Records: []Record{},
				}
				records := zone["records"].([]interface{})
				for _, recordData := range records {
					record := recordData.(map[string]interface{})
					r := Record{
						Name:  record["name"].(string),
						Type:  record["type"].(string),
						Value: record["value"].(string),
					}
					newZone.Records = append(newZone.Records, r)
				}
				allZonesNoHosts = append(allZonesNoHosts, newZone)
			}
		}
	}

	allZones := []Zone{}
	for _, zone := range allZonesNoHosts {
		for _, ip := range allIpAddresses {
			if strings.HasSuffix(ip.Dns, zone.Name) {
				hostName := strings.Replace(ip.Dns, "."+zone.Name, "", -1)
				r := Record{
					Name:  hostName,
					Value: ip.Address.String(),
				}

				if strings.Contains(r.Value, ":") {
					r.Type = "AAAA"
				} else {
					r.Type = "A"
				}

				zone.Records = append(zone.Records, r)
			}
		}
		allZones = append(allZones, zone)
	}

	err = common.CreateDirIfNotExists(g.out)
	if err != nil {
		return fmt.Errorf("CreateDirIfNotExists: %v\n", err)
	}

	for _, zone := range allZones {
		t, err := template.New("forwardDnsZoneTemplate").Parse(forwardDnsZoneTemplate)
		if err != nil {
			return fmt.Errorf("template.New: %v", err)
		}

		p := dnsZoneParams{
			Name:    zone.Name,
			Serial:  serial,
			Records: zone.Records,
		}

		fname := g.out + "/db." + zone.Name
		fd, err := os.Create(fname + ".new")
		if err != nil {
			return fmt.Errorf("os.Open: %v", err)
		}
		defer func() {
			fd.Close()
			os.Rename(fname+".new", fname)
			fmt.Printf("[+] Wrote %s\n", fname)
		}()

		err = t.Execute(fd, p)
		if err != nil {
			return fmt.Errorf("t.Execute: %v", err)
		}
	}

	return nil
}

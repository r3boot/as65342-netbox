package generator

import (
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/r3boot/as65342-netbox/lib/common"
)

const icinga2Template = `#
# Generated on xxx by yyy
#
{{- range .Tenants }}
template Host "{{ .Slug }}-gateway" { 
  max_check_attempts = 3
  check_interval = 1m
  retry_interval = 30s

  check_command = "hostalive"
} 

template Host "{{ .Slug }}-host" { 
  max_check_attempts = 3
  check_interval = 1m
  retry_interval = 30s

  check_command = "hostalive"
} 

template Service "{{ .Slug }}-service" { 
  max_check_attempts = 3
  check_interval = 1m
  retry_interval = 30s
} 

object HostGroup "{{ .Slug }}-servers" {
  display_name = "{{ .Name }} Servers"
  assign where host.vars.tenant == "{{ .Slug }}"
}

{{- end }}

{{- range .Platforms }}
object HostGroup "{{ . }}-servers" {
  display_name = "{{ . }} Servers"
  assign where host.vars.platform == "{{ . }}"
}

{{- end }}

{{- range .Sites }}
object HostGroup "site-{{ . }}-servers" {
  display_name = "{{ . }} Servers"
  assign where host.vars.site == "{{ . }}"
}

{{- end }}

{{- range .Ipv4Gateways }}
object Host "gw-{{ .PrintableAddress }}" {
  import "{{ .Tenant }}-gateway"
  address = "{{ .Address }}"
  display_name = "gw-{{ .PrintableAddress }}"
}

apply Dependency "host-to-gw-{{ .PrintableAddress }}" to Host {
  parent_host_name = "gw-{{ .PrintableAddress }}"
  disable_checks = true
  disable_notifications = true

  assign where host.vars.network == "net-{{ .PrintableNetwork }}"
}

apply Dependency "service-to-gw-{{ .PrintableAddress }}" to Service {
  parent_host_name = "gw-{{ .PrintableAddress }}"
  parent_service_name = "ping4"

  disable_checks = true
  disable_notifications = true

  assign where host.vars.network == "net-{{ .PrintableNetwork }}"
}

{{- end }}

{{- range .Ipv6Gateways }}
object Host "gw-{{ .PrintableAddress }}" {
  import "{{ .Tenant }}-gateway"
  address6 = "{{ .Address }}"
  display_name = "gw-{{ .PrintableAddress }}"
}

apply Dependency "host-to-gw-{{ .PrintableAddress }}" to Host {
  parent_host_name = "gw-{{ .PrintableAddress }}"
  disable_checks = true
  disable_notifications = true

  assign where host.vars.network6 == "net-{{ .PrintableNetwork }}"
}

apply Dependency "service-to-gw-{{ .PrintableAddress }}" to Service {
  parent_host_name = "gw-{{ .PrintableAddress }}"
  parent_service_name = "ping6"

  disable_checks = true
  disable_notifications = true

  assign where host.vars.network6 == "net-{{ .PrintableNetwork }}"
}

{{- end }}

{{- range .Devices }}
object Host "{{ .Name }}" { 
  import "{{ .Tenant }}-host"

  address6 = "{{ .PrimaryIP6 }}"
  address = "{{ .PrimaryIP4 }}"

  vars.platform = "{{ .Platform }}"
  vars.tenant = "{{ .Tenant }}"
  vars.site = "{{ .Site }}"
  vars.network6 = "net-{{ .PrintablePrimaryNet6 }}"
  vars.network = "net-{{ .PrintablePrimaryNet4 }}"

  vars.http_vhosts["{{ .Name }}"] = {
    http_vhost  = "{{ .Name }}"
    http_port   = 443
    http_ssl    = true
    http_sni    = true
    http_uri    = "/_status"
    http_string = "Active connections"
  }
} 

{{- end }}
`

type icinga2Params struct {
	Devices      []common.ManagedDevice
	Tenants      []common.Tenant
	Ipv4Gateways []common.Gateway
	Ipv6Gateways []common.Gateway
	Platforms    []string
	Sites        []string
}

func (g *Generator) Icinga2Config() error {
	gateways, err := g.client.ListGateways()
	if err != nil {
		return fmt.Errorf("ListGateways: %v", err)
	}

	v4Gateways := []common.Gateway{}
	v6Gateways := []common.Gateway{}
	for _, gw := range gateways {
		if strings.Contains(gw.Address, ":") {
			v6Gateways = append(v6Gateways, gw)
		} else {
			v4Gateways = append(v4Gateways, gw)
		}
	}

	tenants, err := g.client.ListTenants()
	if err != nil {
		return fmt.Errorf("ListTenants: %v", err)
	}

	devices, err := g.client.GetHostList()
	if err != nil {
		return fmt.Errorf("GetHostList: %v", err)
	}

	allPlatforms := []string{}
	for _, device := range devices {
		newPlatform := device.Platform
		is_listed := false
		for _, platform := range allPlatforms {
			if platform == newPlatform {
				is_listed = true
			}
		}
		if !is_listed {
			allPlatforms = append(allPlatforms, newPlatform)
		}
	}

	allSites := []string{}
	for _, device := range devices {
		newSite := device.Site
		is_listed := false
		for _, site := range allSites {
			if site == newSite {
				is_listed = true
			}
		}
		if !is_listed {
			allSites = append(allSites, newSite)
		}
	}

	t, err := template.New("ansibleInventory").Parse(icinga2Template)
	if err != nil {
		return fmt.Errorf("template.New: %v", err)
	}

	p := icinga2Params{
		Devices:      devices,
		Tenants:      tenants,
		Ipv4Gateways: v4Gateways,
		Ipv6Gateways: v6Gateways,
		Platforms:    allPlatforms,
		Sites:        allSites,
	}

	err = common.CreateDirIfNotExists(g.out)
	if err != nil {
		return fmt.Errorf("CreateDirIfNotExists: %v\n", err)
	}

	fd, err := os.Create(g.out + "/generated.conf.new")
	if err != nil {
		return fmt.Errorf("os.Open: %v", err)
	}
	defer func() {
		fd.Close()
		os.Rename(g.out+"/generated.conf.new", g.out+"/generated.conf")
		fmt.Printf("[+] Wrote %s\n", g.out+"/generated.conf")
	}()

	err = t.Execute(fd, p)
	if err != nil {
		return fmt.Errorf("t.Execute: %v", err)
	}

	return nil
}

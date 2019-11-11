package netboxclient

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/r3boot/as65342-netbox/lib/common"
	"github.com/r3boot/as65342-netbox/lib/netbox/client"
	"github.com/r3boot/as65342-netbox/lib/netbox/client/dcim"
	"github.com/r3boot/as65342-netbox/lib/netbox/client/extras"
	"github.com/r3boot/as65342-netbox/lib/netbox/client/ipam"
	"github.com/r3boot/as65342-netbox/lib/netbox/client/tenancy"
	"github.com/r3boot/as65342-netbox/lib/netbox/client/virtualization"
)

type NetboxClient struct {
	api                 *client.NetBox
	token               common.TokenAuth
	limit               int64
	ipamPrefixesList    *ipam.IpamPrefixesListOK
	dcimDevicesList     *dcim.DcimDevicesListOK
	virtualMachinesList *virtualization.VirtualizationVirtualMachinesListOK
	configContextList   *extras.ExtrasConfigContextListOK
	tenantList          *tenancy.TenancyTenantsListOK
}

func NewNetboxClient(api *client.NetBox, token common.TokenAuth, limit int64) (client *NetboxClient, err error) {
	client = &NetboxClient{
		api:   api,
		token: token,
		limit: limit,
	}

	err = client.PrewarmCaches()
	if err != nil {
		return nil, fmt.Errorf("client.PrewarmCaches: %v\n", err)
	}

	return client, nil
}

func (c *NetboxClient) PrewarmCaches() (err error) {
	c.ipamPrefixesList, err = c.api.Ipam.IpamPrefixesList(&ipam.IpamPrefixesListParams{
		Limit:   &c.limit,
		Context: context.Background(),
	}, c.token)
	if err != nil {
		return fmt.Errorf("IPAM.IPAMPrefixesList: %v", err)
	}

	c.dcimDevicesList, err = c.api.Dcim.DcimDevicesList(&dcim.DcimDevicesListParams{
		Limit:   &c.limit,
		Context: context.Background(),
	}, c.token)
	if err != nil {
		return fmt.Errorf("Dcim.DcimDevicesList: %v", err)
	}

	c.virtualMachinesList, err = c.api.Virtualization.VirtualizationVirtualMachinesList(&virtualization.VirtualizationVirtualMachinesListParams{
		Limit:   &c.limit,
		Context: context.Background(),
	}, c.token)
	if err != nil {
		return fmt.Errorf("Virtualization.VirtualizationVirtualMachinesList: %v", err)
	}

	c.configContextList, err = c.api.Extras.ExtrasConfigContextList(&extras.ExtrasConfigContextListParams{
		Limit:   &c.limit,
		Context: context.Background(),
	}, c.token)
	if err != nil {
		return fmt.Errorf("Extras.ExtrasConfigContextList: %v", err)
	}

	c.tenantList, err = c.api.Tenancy.TenancyTenantsList(&tenancy.TenancyTenantsListParams{
		Limit:   &c.limit,
		Context: context.Background(),
	}, c.token)
	if err != nil {
		return fmt.Errorf("Tenancy.TenancyTenantsList: %v", err)
	}

	return nil
}

func (c *NetboxClient) GetPrefixList(tenant string) (allPrefixes []*net.IPNet, err error) {
	for _, entry := range c.ipamPrefixesList.Payload.Results {
		if !common.IsAllowedTenant(*entry.Tenant.Slug) {
			continue
		}

		_, network, err := net.ParseCIDR(*entry.Prefix)
		if err != nil {
			fmt.Printf("Failed to parse ip: %v\n", *entry.Prefix)
		}

		if len(network.IP) == 16 {
			if size, _ := network.Mask.Size(); size != 64 {
				continue
			}
		} else {
			if size, _ := network.Mask.Size(); size != 24 {
				continue
			}
		}

		allPrefixes = append(allPrefixes, network)
	}

	return allPrefixes, nil
}

func (c *NetboxClient) GetHostList() (allDevices []common.ManagedDevice, err error) {
	// Get details for physical systems
	for _, entry := range c.dcimDevicesList.Payload.Results {
		if !common.IsAllowedTenant(*entry.Tenant.Slug) {
			continue
		}

		if !common.IsAllowedStatus(*entry.Status.Label) {
			continue
		}

		if entry.Platform == nil {
			continue
		}

		if !common.IsAllowedPlatform(*entry.Platform.Slug) {
			continue
		}

		device := common.ManagedDevice{
			Name:     entry.Name,
			Tags:     entry.Tags,
			Tenant:   *entry.Tenant.Slug,
			Platform: *entry.Platform.Slug,
			Config:   entry.ConfigContext,
		}

		device.PrimaryIP, device.PrimaryNet, err = net.ParseCIDR(*entry.PrimaryIP.Address)
		if err != nil {
			return nil, fmt.Errorf("net.ParseCIDR: %v", err)
		}
		device.PrintablePrimaryNet = device.PrimaryNet.IP.String()
		device.PrintablePrimaryNet = strings.Replace(device.PrintablePrimaryNet, "/24", "", -1)
		device.PrintablePrimaryNet = strings.Replace(device.PrintablePrimaryNet, ".", "-", -1)

		device.PrimaryIP4, device.PrimaryNet4, err = net.ParseCIDR(*entry.PrimaryIp4.Address)
		if err != nil {
			return nil, fmt.Errorf("net.ParseCIDR: %v", err)
		}
		device.PrintablePrimaryNet4 = device.PrimaryNet4.IP.String()
		device.PrintablePrimaryNet4 = strings.Replace(device.PrintablePrimaryNet4, "/24", "", -1)
		device.PrintablePrimaryNet4 = strings.Replace(device.PrintablePrimaryNet4, ".", "-", -1)

		device.PrimaryIP6, device.PrimaryNet6, err = net.ParseCIDR(*entry.PrimaryIp6.Address)
		if err != nil {
			return nil, fmt.Errorf("net.ParseCIDR: %v", err)
		}
		device.PrintablePrimaryNet6 = device.PrimaryNet6.IP.String()
		device.PrintablePrimaryNet6 = strings.Replace(device.PrintablePrimaryNet6, "::", "", -1)
		device.PrintablePrimaryNet6 = strings.Replace(device.PrintablePrimaryNet6, ":", "-", -1)

		allDevices = append(allDevices, device)
	}

	// Get details for virtual machines
	for _, entry := range c.virtualMachinesList.Payload.Results {
		if !common.IsAllowedTenant(*entry.Tenant.Slug) {
			continue
		}

		if !common.IsAllowedStatus(*entry.Status.Label) {
			continue
		}

		if entry.Platform == nil {
			continue
		}

		if !common.IsAllowedPlatform(*entry.Platform.Slug) {
			continue
		}

		device := common.ManagedDevice{
			Name:     *entry.Name,
			Tags:     entry.Tags,
			Tenant:   *entry.Tenant.Slug,
			Platform: *entry.Platform.Slug,
			Config:   entry.ConfigContext,
		}

		device.PrimaryIP, device.PrimaryNet, err = net.ParseCIDR(*entry.PrimaryIP.Address)
		if err != nil {
			return nil, fmt.Errorf("net.ParseCIDR: %v", err)
		}
		device.PrintablePrimaryNet = device.PrimaryNet.IP.String()
		device.PrintablePrimaryNet = strings.Replace(device.PrintablePrimaryNet, "/24", "", -1)
		device.PrintablePrimaryNet = strings.Replace(device.PrintablePrimaryNet, ".", "-", -1)

		device.PrimaryIP4, device.PrimaryNet4, err = net.ParseCIDR(*entry.PrimaryIp4.Address)
		if err != nil {
			return nil, fmt.Errorf("net.ParseCIDR: %v", err)
		}
		device.PrintablePrimaryNet4 = device.PrimaryNet4.IP.String()
		device.PrintablePrimaryNet4 = strings.Replace(device.PrintablePrimaryNet4, "/24", "", -1)
		device.PrintablePrimaryNet4 = strings.Replace(device.PrintablePrimaryNet4, ".", "-", -1)

		device.PrimaryIP6, device.PrimaryNet6, err = net.ParseCIDR(*entry.PrimaryIp6.Address)
		if err != nil {
			return nil, fmt.Errorf("net.ParseCIDR: %v", err)
		}
		device.PrintablePrimaryNet6 = device.PrimaryNet6.IP.String()
		device.PrintablePrimaryNet6 = strings.Replace(device.PrintablePrimaryNet6, "::", "", -1)
		device.PrintablePrimaryNet6 = strings.Replace(device.PrintablePrimaryNet6, ":", "-", -1)

		allDevices = append(allDevices, device)
	}

	return allDevices, nil
}

func (c *NetboxClient) ListConfigContexts() (contexts []common.ConfigContext, err error) {
	for _, entry := range c.configContextList.Payload.Results {
		context := common.ConfigContext{
			Name:   *entry.Name,
			Config: entry.Data,
		}
		contexts = append(contexts, context)
	}

	return contexts, nil
}

func (c *NetboxClient) ListTenants() (tenants []common.Tenant, err error) {
	for _, entry := range c.tenantList.Payload.Results {
		tenant := common.Tenant{
			Name: *entry.Name,
			Slug: *entry.Slug,
		}
		tenants = append(tenants, tenant)
	}

	return tenants, nil
}

func (c *NetboxClient) ListGateways() (gateways []common.Gateway, err error) {
	for _, entry := range c.dcimDevicesList.Payload.Results {
		tenant := *entry.Tenant.Slug
		if entry.PrimaryIp4 != nil {
			_, primary_net4, err := net.ParseCIDR(*entry.PrimaryIp4.Address)
			if err != nil {
				return nil, fmt.Errorf("net.ParseCIDR: %v", err)
			}

			size, _ := primary_net4.Mask.Size()
			if size <= 24 {
				newgw := common.Gateway{
					Network: primary_net4.String(),
					Tenant:  tenant,
				}
				primary_net4.IP[3]++
				newgw.Address = primary_net4.IP.String()
				newgw.PrintableAddress = strings.Replace(newgw.Address, ".", "-", -1)
				newgw.PrintableNetwork = strings.Replace(newgw.Network, "/24", "", -1)
				newgw.PrintableNetwork = strings.Replace(newgw.PrintableNetwork, ".", "-", -1)

				is_listed := false
				for _, gw := range gateways {
					if gw.Address == newgw.Address {
						is_listed = true
					}
				}

				if !is_listed {
					gateways = append(gateways, newgw)
				}
			}
		}

		if entry.PrimaryIp6 != nil {
			_, primary_net6, err := net.ParseCIDR(*entry.PrimaryIp6.Address)
			if err != nil {
				return nil, fmt.Errorf("net.ParseCIDR: %v", err)
			}

			size, _ := primary_net6.Mask.Size()
			if size <= 64 {
				newgw := common.Gateway{
					Network: primary_net6.String(),
					Tenant:  tenant,
				}
				primary_net6.IP[15]++
				newgw.Address = primary_net6.IP.String()
				newgw.PrintableAddress = strings.Replace(newgw.Address, ":", "-", -1)
				newgw.PrintableNetwork = strings.Replace(newgw.Network, "::/64", "", -1)
				newgw.PrintableNetwork = strings.Replace(newgw.PrintableNetwork, ":", "-", -1)

				is_listed := false
				for _, gw := range gateways {
					if gw.Address == newgw.Address {
						is_listed = true
					}
				}

				if !is_listed {
					gateways = append(gateways, newgw)
				}
			}
		}
	}

	for _, entry := range c.virtualMachinesList.Payload.Results {
		tenant := *entry.Tenant.Slug
		if entry.PrimaryIp4 != nil {
			_, primary_net4, err := net.ParseCIDR(*entry.PrimaryIp4.Address)
			if err != nil {
				return nil, fmt.Errorf("net.ParseCIDR: %v", err)
			}

			size, _ := primary_net4.Mask.Size()
			if size <= 24 {
				newgw := common.Gateway{
					Network: primary_net4.String(),
					Tenant:  tenant,
				}
				primary_net4.IP[3]++
				newgw.Address = primary_net4.IP.String()
				newgw.PrintableAddress = strings.Replace(newgw.Address, ".", "-", -1)
				newgw.PrintableNetwork = strings.Replace(newgw.Network, "/24", "", -1)
				newgw.PrintableNetwork = strings.Replace(newgw.PrintableNetwork, ".", "-", -1)

				is_listed := false
				for _, gw := range gateways {
					if gw.Address == newgw.Address {
						is_listed = true
					}
				}

				if !is_listed {
					gateways = append(gateways, newgw)
				}
			}
		}

		if entry.PrimaryIp6 != nil {
			_, primary_net6, err := net.ParseCIDR(*entry.PrimaryIp6.Address)
			if err != nil {
				return nil, fmt.Errorf("net.ParseCIDR: %v", err)
			}

			size, _ := primary_net6.Mask.Size()
			if size <= 64 {
				newgw := common.Gateway{
					Network: primary_net6.String(),
					Tenant:  tenant,
				}
				primary_net6.IP[15]++
				newgw.Address = primary_net6.IP.String()
				newgw.PrintableAddress = strings.Replace(newgw.Address, ":", "-", -1)
				newgw.PrintableNetwork = strings.Replace(newgw.Network, "::/64", "", -1)
				newgw.PrintableNetwork = strings.Replace(newgw.PrintableNetwork, ":", "-", -1)

				is_listed := false
				for _, gw := range gateways {
					if gw.Address == newgw.Address {
						is_listed = true
					}
				}

				if !is_listed {
					gateways = append(gateways, newgw)
				}
			}
		}
	}

	return gateways, nil
}

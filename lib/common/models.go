package common

import "net"

type ManagedDevice struct {
	Name                 string
	PrimaryIP            net.IP
	PrimaryNet           *net.IPNet
	PrintablePrimaryNet  string
	PrimaryIP4           net.IP
	PrimaryNet4          *net.IPNet
	PrintablePrimaryNet4 string
	PrimaryIP6           net.IP
	PrimaryNet6          *net.IPNet
	PrintablePrimaryNet6 string
	Platform             string
	Site                 string
	Tenant               string
	Tags                 []string
	Config               interface{}
}

type ConfigContext struct {
	Name   string
	Config interface{}
}

type Tenant struct {
	Name string
	Slug string
}

type Gateway struct {
	Address          string
	Network          string
	Tenant           string
	PrintableAddress string
	PrintableNetwork string
}

type IpAddress struct {
	Address net.IP
	Network *net.IPNet
	Dns     string
	Tenant  string
}

package generator

import (
	"fmt"

	"github.com/r3boot/as65342-netbox/lib/common"
)

type Zone struct {
	Name    string
	Devices []common.ManagedDevice
}

func (g Generator) ReverseDNS() error {
	allPrefixes, err := g.client.GetPrefixList("as65342")
	if err != nil {
		return fmt.Errorf("GetPrefixList: %v", err)
	}

	devices, err := g.client.GetHostList()
	if err != nil {
		return fmt.Errorf("GetHostList: %v", err)
	}

	allZones := []Zone{}
	for _, prefix := range allPrefixes {
	}

}

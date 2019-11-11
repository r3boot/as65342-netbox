package generator

import (
	"fmt"
	"os/user"

	"github.com/r3boot/as65342-netbox/lib/netboxclient"
)

type Generator struct {
	client   *netboxclient.NetboxClient
	out      string
	Username string
}

func NewGenerator(c *netboxclient.NetboxClient, output string) (*Generator, error) {
	g := &Generator{
		client: c,
		out:    output,
	}

	u, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("user.Current: %v\n", err)
	}
	if u.Name != "" {
		g.Username = u.Name
	} else {
		g.Username = u.Username
	}

	return g, nil
}

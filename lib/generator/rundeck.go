package generator

import (
	"fmt"
	"os"
	"strings"

	"github.com/r3boot/as65342-netbox/lib/common"
)

func (g Generator) RundeckHosts() error {
	allHosts, err := g.client.GetHostList()
	if err != nil {
		return fmt.Errorf("GetHostList: %v", err)
	}

	err = common.CreateDirIfNotExists(g.out)
	if err != nil {
		return fmt.Errorf("CreateDirIfNotExists: %v\n", err)
	}

	fname := g.out + "/hosts.yml"
	fd, err := os.Create(fname + ".new")
	if err != nil {
		return fmt.Errorf("os.Open: %v", err)
	}
	defer func() {
		fd.Close()
		os.Rename(fname+".new", fname)
		fmt.Printf("[+] Wrote %s\n", fname)
	}()

	for _, host := range allHosts {
		username := "rundeck"
		os := "linux"
		switch host.Platform {
		case "coreos":
			{
				username = "core"
			}
		case "openbsd":
			{
				os = "bsd"
			}
		}

		line := fmt.Sprintf("%s:\n", strings.Split(host.Name, ".")[0])
		line += fmt.Sprintf("  hostname: %s\n", host.Name)
		line += fmt.Sprintf("  username: %s\n", username)
		line += fmt.Sprintf("  osFamily: %s\n", os)
		line += fmt.Sprintf("  osName: %s\n", host.Platform)

		fd.Write([]byte(line))
	}

	return nil
}

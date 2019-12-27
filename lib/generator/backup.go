package generator

import (
	"fmt"
	"os"

	"github.com/r3boot/as65342-netbox/lib/common"
)

func (g Generator) BackupHosts() error {
	allHosts, err := g.client.GetHostList()
	if err != nil {
		return fmt.Errorf("GetHostList: %v", err)
	}

	err = common.CreateDirIfNotExists(g.out)
	if err != nil {
		return fmt.Errorf("CreateDirIfNotExists: %v\n", err)
	}

	fname := g.out + "/backup.hosts"
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
		line := fmt.Sprintf("%s,%s,%s\n", host.Name, host.Platform, host.Site)
		fd.Write([]byte(line))
	}

	return nil
}

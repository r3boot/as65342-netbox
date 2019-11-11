package common

import (
	"fmt"
	"net"
	"os"
	"strings"
)

var (
	allowedTenants   []string = []string{"as65342"}
	allowedPlatforms []string = []string{"centos", "openbsd"}
	allowedStatus    []string = []string{"Active"}
)

func IsAllowedTenant(tenant string) bool {
	for idx := range allowedTenants {
		if allowedTenants[idx] == tenant {
			return true
		}
	}
	return false
}

func IsAllowedStatus(status string) bool {
	for idx := range allowedStatus {
		if allowedStatus[idx] == status {
			return true
		}
	}
	return false
}

func IsAllowedPlatform(platform string) bool {
	for idx := range allowedPlatforms {
		if allowedPlatforms[idx] == platform {
			return true
		}
	}
	return false
}

func CreateDirIfNotExists(dirName string) error {
	fd, err := os.Stat(dirName)
	if err == nil {
		if !fd.IsDir() {
			return fmt.Errorf("fd.IsDir: not a directory")
		} else {
			return nil
		}
	}

	err = os.Mkdir(dirName, 0755)
	if err != nil {
		return fmt.Errorf("os.Mkdir: %v", err)
	}

	return nil
}

func ToDnsZoneName(network *net.IPNet) string {
	i := network.IP
	result := ""
	if len(i) == 16 {
		tmp := fmt.Sprintf("%02x%02x%02x%02x%02x%02x%02x%02x", i[0], i[1], i[2], i[3], i[4], i[5], i[6], i[7])
		for _, v := range tmp {
			if result == "" {
				result = string(v)
			} else {
				result = string(v) + "." + result
			}
		}
		result = result + ".ip6.arpa"
	} else {
		result = fmt.Sprintf("%d.%d.%d.in-addr.arpa", i[2], i[1], i[0])
	}

	return result
}

func ToFqdn(name string) string {
	if !strings.HasSuffix(name, ".") {
		return name + "."
	}
	return name
}

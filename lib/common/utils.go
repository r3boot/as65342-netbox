package common

import (
	"fmt"
	"os"
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
			return fmt.Errorf("fd.Isdir: not a directory")
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

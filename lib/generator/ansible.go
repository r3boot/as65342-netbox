package generator

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"

	"github.com/r3boot/as65342-netbox/lib/common"
)

const inventoryFileTemplate = `#
# Generated on xxx by yyy
#
[all:vars]
ansible_connection = ssh
ansible_user = r3boot
ansible_become_pass = {{"{{"}} ansible_become_pass {{"}}"}}

[all]
{{- range .Devices }}
{{ .Name }}
{{- end }}
{{ range $tag, $hosts := .Tags }}
[{{ $tag }}]
{{- range $hosts }}
{{ .Name }}
{{ end }}
{{- end }}
`

type inventoryFileParams struct {
	Devices []common.ManagedDevice
	Tags    map[string][]common.ManagedDevice
}

func (g *Generator) AnsibleInventory() (err error) {
	entries, err := g.client.GetHostList()
	if err != nil {
		return fmt.Errorf("client.GetHostListFor: %v", err)
	}

	tags := make(map[string][]common.ManagedDevice)
	for _, entry := range entries {
		for _, tag := range entry.Tags {
			tags[tag] = append(tags[tag], entry)
		}
	}

	t, err := template.New("ansibleInventory").Parse(inventoryFileTemplate)
	if err != nil {
		return fmt.Errorf("template.New: %v", err)
	}

	p := inventoryFileParams{
		Devices: entries,
		Tags:    tags,
	}

	err = common.CreateDirIfNotExists(g.out)
	if err != nil {
		return fmt.Errorf("CreateDirIfNotExists: %v\n", err)
	}

	fd, err := os.Create(g.out + "/hosts.new")
	if err != nil {
		return fmt.Errorf("os.Open: %v", err)
	}
	defer func() {
		fd.Close()
		os.Rename(g.out+"/hosts.new", g.out+"/hosts")
		fmt.Printf("[+] Wrote %s\n", g.out+"/hosts")
	}()

	err = t.Execute(fd, p)
	if err != nil {
		return fmt.Errorf("t.Execute: %v", err)
	}

	return nil
}

func (g *Generator) AnsibleGroupVars() error {
	entries, err := g.client.ListConfigContexts()
	if err != nil {
		return fmt.Errorf("client.GetHostListFor: %v", err)
	}

	err = common.CreateDirIfNotExists(g.out + "/group_vars")
	if err != nil {
		return fmt.Errorf("CreateDirIfNotExists: %v\n", err)
	}

	for _, entry := range entries {
		fname := entry.Name + ".json"
		fullFname := g.out + "/group_vars/" + fname

		fd, err := os.Create(fullFname + ".new")
		if err != nil {
			return fmt.Errorf("os.Open: %v", err)
		}
		defer func() {
			fd.Close()
			os.Rename(fullFname+".new", fullFname)
			fmt.Printf("[+] Wrote %s\n", fullFname)
		}()

		data, err := json.Marshal(entry.Config)
		if err != nil {
			return fmt.Errorf("json.Marshal: %v", err)
		}

		fd.Write(data)
	}

	return nil
}

func (g *Generator) AnsibleHostVars() error {
	entries, err := g.client.GetHostList()
	if err != nil {
		return fmt.Errorf("client.GetHostListFor: %v", err)
	}

	err = common.CreateDirIfNotExists(g.out + "/host_vars")
	if err != nil {
		return fmt.Errorf("CreateDirIfNotExists: %v\n", err)
	}

	for _, entry := range entries {
		fname := entry.Name + ".json"
		fullFname := g.out + "/host_vars/" + fname

		fd, err := os.Create(fullFname + ".new")
		if err != nil {
			return fmt.Errorf("os.Open: %v", err)
		}
		defer func() {
			fd.Close()
			os.Rename(fullFname+".new", fullFname)
			fmt.Printf("[+] Wrote %s\n", fullFname)
		}()

		data, err := json.Marshal(entry.Config)
		if err != nil {
			return fmt.Errorf("json.Marshal: %v", err)
		}

		fd.Write(data)
	}
	return nil
}
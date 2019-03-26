package templategen

import (
	"bufio"
	"os"
	"text/template"
)

// JunosTemplatePayload holds data for populating Junos device configs
type JunosTemplatePayload struct {
	Gateway    string
	FixedIP    string
	HostName   string
	DomainName string
	DNSServers []string
	NTPServers []string
}

// SaveJunosConfig does as it says on the tin!
func SaveJunosConfig(file string, templ string, payload JunosTemplatePayload) error {

	t := template.Must(template.New("junos.template").ParseFiles(templ))

	// We don't care about erroring out. Error out silently here.
	os.Remove(file)

	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return err
	}

	w := bufio.NewWriter(f)

	err = t.Execute(w, payload)

	err = w.Flush()

	if err := f.Close(); err != nil {
		return err
	}

	return nil
}

// Protected by BSD 3 clause license
// Package implements the core config logic for EZJunosZTP

package cfg

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"reflect"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	rt "github.com/networkbootstrap/ztpmanagercode/roottypes"
	templategen "github.com/networkbootstrap/ztpmanagercode/templategen/junos"
)

const ifacetmpl = `# /etc/default/isc-dhcp-server
#
# Generated: by "EZJunosZTP" Tool
#
# Timestamp: %s
# 
# On what interfaces should the DHCP server (dhcpd) serve DHCP requests?
#     Separate multiple interfaces with spaces, e.g. "eth0 eth1".

INTERFACESv4="%s"
`

const dhcpdtmpl = `# /etc/dhcpd/dhcpd.conf
# dhcpd.conf
#
# Generated: by "EZJunosZTP" Tool
#
# Timestamp: %s

# The ddns-updates-style parameter controls
ddns-update-style none;

# If this DHCP server is the official DHCP server for the local
# network, the authoritative directive should be uncommented.
authoritative;

# Use this to send dhcp log messages to a different log file (you also
# have to hack syslog.conf to complete the redirection).
log-facility local7;

# Junos ZTP options
option space ezjunosztp;
option ezjunosztp.image-file-name code 0 = text;
option ezjunosztp.config-file-name code 1 = text;
option ezjunosztp.image-file-type code 2 = text;
option ezjunosztp.transfer-mode code 3 = text;
option ezjunosztp-encap code 43 = encapsulate ezjunosztp;
option ezjunosztp-file-server code 150 = ip-address;
`

const subnettmpl = `# Subnet definition
subnet %s netmask %s {
	range dynamic-bootp %s %s;
	option routers %s;
}
`

const hosttmpl = `	# Host definition
	host %s.%s {
		hardware ethernet %s;
		fixed-address %s;
		option host-name "%s";
`

// Cfg type holds core DHCP ISC configuration
type Cfg struct {
	Core  CoreCfg             `json:"core"`
	Hosts map[string]rt.Hosts `json:"hosts"`
}

// CoreCfg holds core info
type CoreCfg struct {
	HTTPUser            string   `json:"httpuser"`
	HTTPPasswd          string   `json:"httppasswd"`
	ServerURL           string   `json:"serverurl"`
	ServerPort          int      `json:"srverport"`
	HTTPConfigsLocation string   `json:"-"` // Directory for serving configurations "configs"
	HTTPImagesLocation  string   `json:"-"` // Directory for serving configurations "images"
	FileConfigsLocation string   `json:"-"` // Directory for generating configurations "./configs"
	FileImagesLocation  string   `json:"-"` // Directory for generating configurations "./configs"
	DHCPDPath           string   `json:"-"` // /etc/dhcpd/dhcpd.conf
	DHCPPath            string   `json:"-"` // /etc/default/isc-dhcp-server
	DHCPIface           string   `json:"dhcpiface" dhcpd:"INTERFACESv4"`
	DomainName          string   `json:"domainname" dhcpd:"option domain-name "`
	DNSServers          []string `json:"dnservers" dhcpd:"option domain-name-servers"`
	DefaultLease        int      `json:"defaultlease" dhcpd:"default-lease-time "`
	MaxLease            int      `json:"maxlease" dhcpd:"max-lease-time "`
	Subnet              string   `json:"subnet" dhcpd:"subnet"`
	SubnetMask          string   `json:"subnetmask" dhcpd:"netmask"`
	NonCfgRangeLow      string   `json:"noncfgrangelow"`
	NonCfgRangeHigh     string   `json:"noncfgrangehigh"`
	SubnetRouter        string   `json:"subnetrouter"  dhcpd:"option routers"`
	TransferMode        string   `json:"transfermode" dhcpd:"option ezjunosztp.transfer-mode"`
	FileServer          string   `json:"fileserver" dhcpd:"option ezjunosztp-file-server"`
	NTPServers          []string `json:"ntpservers" dhcpd:"option ntp-servers"`
}

// NewCfg returns a new empty Cfg struct
func NewCfg() Cfg {
	rtn := Cfg{}
	// Nil pointer panic error bug. Sometimes the config file is empty, so this never gets initd.
	rtn.Hosts = make(map[string]rt.Hosts)
	return rtn
}

// Parse unmarshalls the TOML based configuration text file on to c
func (c *Cfg) Parse(cfgfile string) error {
	if _, err := toml.DecodeFile(cfgfile, c); err != nil {
		fmt.Print("Error detected\n")
		return err
	}

	return nil
}

// Save marshals and saves the content of c
// The REST API handler is blocking (Go routine single buffered channel), so no need to block here
//func (c *Cfg) Save(cfgfile string, send chan rt.Envelope) error {
func (c *Cfg) Save(cfgfile string) error {

	// We need to quickly load up the FixedIP address fields, ConfigLocations and image files
	for k, v := range c.Hosts {
		v.FixedIP = k
		v.CfgFile = c.Core.HTTPConfigsLocation + "/" + v.HostName + ".conf"
		if v.CfgImage != "" {
			tmpImage := v.CfgImage
			v.CfgImage = c.Core.HTTPImagesLocation + "/" + tmpImage
		}
		c.Hosts[k] = v
	}

	os.Remove(cfgfile)
	f, err := os.OpenFile(cfgfile, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}

	w := bufio.NewWriter(f)
	tomlEncoder := toml.NewEncoder(w)

	err = tomlEncoder.Encode(c)
	if err != nil {
		return err
	}

	err = w.Flush()

	if err := f.Close(); err != nil {
		return err
	}

	return nil
}

// CreateIfaceSetting is a func that returns a stringified version of the cache for the DHCPd configuration
// func (c Cfg) CreateIfaceSetting(send chan rt.Envelope) (string, error) {
func (c Cfg) CreateIfaceSetting() (string, error) {
	// Get timestamp
	timeStamp := time.Now()
	strtimeStamp := timeStamp.String()

	buf := new(bytes.Buffer)

	buf.Write([]byte(fmt.Sprintf(ifacetmpl, strtimeStamp, c.Core.DHCPIface)))
	buf.Write([]byte(fmt.Sprint("\n")))

	return buf.String(), nil
}

// CreateDHCPd is a func that returns a stringified version of the cache for the DHCPd configuration
func (c Cfg) CreateDHCPd() (string, error) {
	// Get timestamp
	timeStamp := time.Now()
	strtimeStamp := timeStamp.String()

	buf := new(bytes.Buffer)
	buf.Write([]byte(fmt.Sprintf(dhcpdtmpl, strtimeStamp)))
	buf.Write([]byte(fmt.Sprint("\n")))

	// Let's deal with the core
	core := c.Core
	tcore := reflect.TypeOf(core)

	field, _ := tcore.FieldByName("DomainName")
	buf.Write([]byte(field.Tag.Get("dhcpd")))
	buf.Write([]byte(fmt.Sprintf("\"%s\";\n", c.Core.DomainName)))

	field, _ = tcore.FieldByName("DNSServers")
	buf.Write([]byte(field.Tag.Get("dhcpd")))
	for k, v := range c.Core.DNSServers {
		if k == 0 {
			buf.Write([]byte(fmt.Sprintf(" %s", v)))
		} else {
			buf.Write([]byte(fmt.Sprintf(", %s", v)))
		}
	}
	buf.Write([]byte(fmt.Sprint(";\n")))

	field, _ = tcore.FieldByName("DefaultLease")
	buf.Write([]byte(field.Tag.Get("dhcpd")))
	buf.Write([]byte(fmt.Sprintf("%v;\n", c.Core.DefaultLease)))

	field, _ = tcore.FieldByName("MaxLease")
	buf.Write([]byte(field.Tag.Get("dhcpd")))
	buf.Write([]byte(fmt.Sprintf("%v;\n", c.Core.MaxLease)))

	field, _ = tcore.FieldByName("Subnet")
	subnetStr := c.MakeSubnet()
	buf.Write(subnetStr.Bytes())

	// Deal with Group creation
	buf.Write([]byte("\ngroup {\n"))

	field, _ = tcore.FieldByName("FileServer")
	buf.Write([]byte(fmt.Sprintf("\t%s", field.Tag.Get("dhcpd"))))
	buf.Write([]byte(fmt.Sprintf(" %s;\n", c.Core.FileServer)))

	field, _ = tcore.FieldByName("TransferMode")
	buf.Write([]byte(fmt.Sprintf("\t%s", field.Tag.Get("dhcpd"))))
	buf.Write([]byte(fmt.Sprintf(" \"%s\";\n", c.Core.TransferMode)))

	field, _ = tcore.FieldByName("NTPServers")
	buf.Write([]byte(fmt.Sprintf("\t%s", field.Tag.Get("dhcpd"))))
	for k, v := range c.Core.NTPServers {
		if k == 0 {
			buf.Write([]byte(fmt.Sprintf(" %s", v)))
		} else {
			buf.Write([]byte(fmt.Sprintf(", %s", v)))
		}
	}
	buf.Write([]byte(fmt.Sprint(";\n")))

	// Now deal with host creation. We must indent by 2x tabs here.
	for _, v := range c.Hosts {
		buf.Write(MakeHost(v, c.Core.DomainName, c.Core.HTTPConfigsLocation, c.Core.HTTPImagesLocation).Bytes())
	}
	// End Group creation
	buf.Write([]byte("}\n"))

	return buf.String(), nil
}

// MakeSubnet returns a subnet string
func (c Cfg) MakeSubnet() *bytes.Buffer {
	buf := new(bytes.Buffer)
	buf.Write([]byte("\n"))
	buf.Write([]byte(fmt.Sprintf(subnettmpl, c.Core.Subnet, c.Core.SubnetMask, c.Core.NonCfgRangeLow, c.Core.NonCfgRangeHigh, c.Core.SubnetRouter)))
	return buf
}

// MakeHost returns a subnet string
func MakeHost(h rt.Hosts, d string, cdir, imgdir string) *bytes.Buffer {
	buf := new(bytes.Buffer)
	buf.Write([]byte("\n"))

	thost := reflect.TypeOf(h)

	buf.Write([]byte(fmt.Sprintf(hosttmpl, h.HostName, d, h.Ethernet, h.FixedIP, h.HostName)))
	if h.CfgFile != "" {
		field, _ := thost.FieldByName("CfgFile")
		buf.Write([]byte(fmt.Sprintf("\t\t%s", field.Tag.Get("dhcpd"))))
		buf.Write([]byte(fmt.Sprintf("\"%s\";\n", h.CfgFile)))
	}
	if h.CfgImage != "" {
		field, _ := thost.FieldByName("CfgImage")
		buf.Write([]byte(fmt.Sprintf("\t\t%s ", field.Tag.Get("dhcpd"))))
		buf.Write([]byte(fmt.Sprintf("\"%s\";\n", h.CfgImage)))
	}
	buf.Write([]byte("\t}\n"))
	return buf
}

// String gets the map of hosts and combines it, then returns the data
func (c *Cfg) String(send chan rt.Envelope) (string, error) {
	var b []byte
	buf := bytes.NewBuffer(b)
	tomlEncoder := toml.NewEncoder(buf)
	err := tomlEncoder.Encode(c.Core)
	if err != nil {
		return "", nil
	}
	returnString := buf.String()

	// Now let's get the stringified version of our TOML from the Hosts map
	reqChan := make(chan rt.Envelope, 1)
	req := rt.Envelope{}
	req.CRUD = rt.STRING
	req.Response = reqChan
	send <- req
	resp := <-req.Response

	if resp.CRUD == rt.OK {
		returnString += "\n"
		returnString += resp.String
		return returnString, nil
	}
	err = errors.New("Error retrieving stringified map hosts")

	return "", err
}

// APIResponder is a GR that responds to REST API calls
// Note, the cache doesn't need to be locked. If we try, we'll get deadlocked.
func (c *Cfg) APIResponder(cachesend chan rt.Envelope, fname string, wg *sync.WaitGroup) (recvch chan rt.Envelope, finish chan struct{}) {
	recvch = make(chan rt.Envelope, 1)
	finish = make(chan struct{})
	deletelist := []string{}

	go func() {
		for {
			select {
			case recv := <-recvch:

				switch recv.CRUD {
				case rt.DELETE:
					resp := rt.Envelope{}
					hostname := c.Hosts[recv.FixedIP].HostName
					deletelist = append(deletelist, hostname)
					resp.CRUD = rt.OK
					recv.Response <- resp

				case rt.SAVECFG:
					// Delete the entries from the delete list and kill the list
					for _, v := range deletelist {
						filename := fmt.Sprintf("%s/%s.conf", c.Core.FileConfigsLocation, v)

						err := os.Remove(filename)
						if err != nil {
							fmt.Print(err)
						}
					}
					// Kill deletelist and rebirth
					deletelist = []string{}

					resp := rt.Envelope{}
					err := c.Save(fname)
					if err != nil {
						fmt.Print(err)
						resp.CRUD = rt.ERROR
						recv.Response <- resp
						// We're done here
						break
					}

					dhcpdStr, err := c.CreateDHCPd()
					if err != nil {
						fmt.Print(err)
						resp.CRUD = rt.ERROR
						recv.Response <- resp
						break
					}

					// Save the contents of the dhcpdStr string to the file for the DHCPD path
					os.Remove(c.Core.DHCPDPath)
					f, err := os.OpenFile(c.Core.DHCPDPath, os.O_RDWR|os.O_CREATE, 0755)
					if err != nil {
						log.Fatal(err)
					}

					w := bufio.NewWriter(f)
					_, err = w.WriteString(dhcpdStr)
					if err != nil {
						fmt.Print(err)
						resp.CRUD = rt.ERROR
						recv.Response <- resp
						break
					}
					err = w.Flush()
					if err := f.Close(); err != nil {
						fmt.Print(err)
						resp.CRUD = rt.ERROR
						recv.Response <- resp
						break
					}

					// Get the dhcp (iface) configs and save
					dhcpStr, err := c.CreateIfaceSetting()
					if err != nil {
						fmt.Print(err)
						resp.CRUD = rt.ERROR
						recv.Response <- resp
						break
					}

					// Save the contents of the dhcpdStr string to the file for the DHCPD path
					os.Remove(c.Core.DHCPPath)
					f, err = os.OpenFile(c.Core.DHCPPath, os.O_RDWR|os.O_CREATE, 0755)
					if err != nil {
						log.Fatal(err)
					}

					w = bufio.NewWriter(f)
					_, err = w.WriteString(dhcpStr)
					if err != nil {
						fmt.Print(err)
						resp.CRUD = rt.ERROR
						recv.Response <- resp
						break
					}
					err = w.Flush()
					if err := f.Close(); err != nil {
						fmt.Print(err)
						resp.CRUD = rt.ERROR
						recv.Response <- resp
						break
					}

					// Now for the fun part, let's generate the device configurations! Whoop whoop.
					// Get device template payload for each device
					for _, v := range c.Hosts {
						tmplPayload := templategen.JunosTemplatePayload{}
						tmplPayload.DNSServers = c.Core.DNSServers
						tmplPayload.DomainName = c.Core.DomainName
						tmplPayload.Gateway = c.Core.SubnetRouter
						tmplPayload.NTPServers = c.Core.NTPServers
						tmplPayload.FixedIP = v.FixedIP
						tmplPayload.HostName = v.HostName

						// Send parameters to method to save file
						cfgFileLoc := c.Core.FileConfigsLocation + "/" + v.HostName + ".conf"

						// If you want to extend and insert another vendor, start here!
						switch v.Vendor {
						case "junos":
							err := templategen.SaveJunosConfig(cfgFileLoc, "./templates/junos/junos.template", tmplPayload)
							if err != nil {
								fmt.Print(err)
								resp.CRUD = rt.ERROR
								recv.Response <- resp
								break
							}
						}
					}

					// Find a way
					go func() {
						cmd := exec.Command("systemctl", "restart", "isc-dhcp-server")
						err := cmd.Run()
						if err != nil {
							fmt.Printf("Issue restarting ISC service: %s \n", err)
						}
					}()

					resp.CRUD = rt.OK

					recv.Response <- resp
				}
			case <-finish:
				// We're finished *sob*
				wg.Done()
				return
			}
		}
	}()
	return
}

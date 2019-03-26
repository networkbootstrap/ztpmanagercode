// Package roottypes contains the things needed to deal with preventing circular dependencies
// Protected by BSD 3 clause license
package roottypes

// Enums for CRUD operations
const (
	// CREATE = create action
	CREATE = iota
	// READHOST = read action for a single host
	READHOST
	// READHOSTS = read the list of hosts currently in the cache
	READHOSTS
	// UPDATE = update action
	UPDATE
	// DELETE = delete action
	DELETE
	// OK = Response encoded
	OK
	// ERROR = oh dear
	ERROR
	// STRING = get in memory DB in text form
	STRING
	// LOCK = locks the cache for reading
	LOCK
	// UNLOCK = unlocks the cache for normal operations
	UNLOCK
	// SAVECFG = saves the actual config in TOML for this application
	SAVECFG
	// SAVEDHCPD = saves the DHCPD config and isc-dhcp config which contains the interface stuffs, it also generates device templates
)

// Envelope gives us a dirty way to use the channel without infecting the Hosts struct with a chan (causes encoding issues).
type Envelope struct {
	Response chan Envelope `json:"-" toml:"-"`
	CRUD     int           `json:"-"`
	String   string        `json:"-" toml:"-"`
	HostList []string      `json:"-" toml:"-"`
	Hosts
}

// Hosts holds data for a single DHCP ISC ZTP host
type Hosts struct {
	Ethernet string `json:"ethernetaddress" dhcpd:"hardware ethernet "`
	FixedIP  string `json:"fixedipaddress" dhcpd:"fixed-address "`
	HostName string `json:"hostname" dhcpd:"option host-name "`
	CfgFile  string `json:"cfgfile" dhcpd:"option ezjunosztp.config-file-name " toml:"-"`
	CfgImage string `json:"imagefile" dhcpd:"option ezjunosztp.image-file-name "`
	UpdateIP string `json:"-" toml:"-"`
	Vendor   string `json:"vendor"`
}

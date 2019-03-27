# Usage

This project is simple to configure and simple to use providing you understand the rules.

- Any change made via the HTTP JSON API must be saved through the API
- If you change the contents of the `config.toml` file, reload the application to force a configuration read
- The webserver running on port 80 will serve the contents of `configs` and `images`. The locations of which are read from the `config.toml`
- Templates for the initial configuration are stored in the `templates/vendor/vendor.template` pattern

## Config.Toml

This file contains the core configuration required to bootstrap the ZTPManager. Host information can also go in here if you're confident enough to create the entries. The JSON API actually drives an operation that does this anyway. To re-word this, if you do a POST operation on `/hosts` then do a POST on `/save`, the application adds the new host to this file, writes the `dhcpd.conf` file and generates the device configuration from the templates in the `templates` directory. What follows is the generic `config.toml` file contents and an explanation of each field.

```bash
[Core]
  HTTPUser = "admin"
  HTTPPasswd = "Passw0rd"
  ServerURL = "localhost"
  ServerPort = 1323
  HTTPConfigsLocation = "configs"
  HTTPImagesLocation = "images"
  FileConfigsLocation = "./configs"
  FileImagesLocation = "./images"
  DHCPDPath = "/etc/dhcp/dhcpd.conf"
  DHCPPath = "/etc/default/isc-dhcp-server"
  DHCPIface = "ens34"
  DomainName = "simpledemo.net"
  DNSServers = ["8.8.8.8", "8.8.4.4"]
  DefaultLease = 600
  MaxLease = 7200
  Subnet = "192.168.50.0"
  SubnetMask = "255.255.255.0"
  NonCfgRangeLow = "192.168.50.20"
  NonCfgRangeHigh = "192.168.50.25"
  SubnetRouter = "192.168.50.1"
  TransferMode = "http"
  FileServer = "192.168.50.254"
  NTPServers = ["192.168.50.254"]

[Hosts]
  [Hosts."192.168.50.100"]
    Ethernet = "00:0c:29:4d:3d:cc"
    FixedIP = "192.168.50.100"
    HostName = "demo01"
    Vendor = "junos"
    CfgImage = ""
```

This configuration file is written in TOML or "Tom's Obvious, Minimal Language", named after Tom Preston-Wener. It is easy to work with and easy to read.

The `[Core]` section contains information relevant to the isc-dhcp-server itself. `[Hosts]` naturally contains information relevant to hosts. Each host entry is a key to a map/dictionary and thus must be unique. If you have a host entry with the same IP being referenced, it will overwrite previous entries. You've been warned!

__HTTPUser__
This is the HTTP basic auth username.

__HTTPPasswd__
This is the HTTP basic auth password.

__ServerURL__
This is the URL; either FQDN or IP address of the server the ZTPManager application will run. 

__ServerPort__
This is the TCP port number the configuration API will run. The file delivery mechanism is hardwired to TCP 80 and is not configurable.

__HTTPConfigsLocation__
This is the name of the configs directory that the webserver will listen for. The webserver maps this variable name to the file path (below).

__FileConfigsLocation__
This is the absolute or relative location of the directory on the system that will serve files.

__HTTPImagesLocation__
This is the name of the images directory that the webserver will listen for. The webserver maps this variable name to the file path (below).

__FileImagesLocation__
This is the absolute or relative location of the directory on the system that will serve files.

__DHCPDPath__
The location of the `dhcpd.conf` file.

__DHCPPath__
The location of the directory which contains the `isc-dhcp-server` file, which contains basic configuration like the interface to bind the dhcp service to.

__DHCPIface__
The name of the interface to bind the dhcp service to.

__DomainName__
Domain name for the DHCP service.

__DNSServers__
List of DNS servers for the DHCP service.

__DefaultLease__
Default lease time for DHCP leases.

__MaxLease__
Maximum lease time for DHCP leases.

__Subnet__
Subnet from which leases are made.

__SubnetMask__
Mask for the subnet.

__NonCfgRangeLow__
Some devices just need DHCP without the ZTP aspect. This range is for those devices. This is the low address for the pool.

__NonCfgRangeHigh__
This is the high address for the pool (see above).

__SubnetRouter__
Router for the subnet.

__FileServer__
The address of the file server. This should be the IP address of the server interface that serves DHCP. At some point, this will be automatic. Today it is not.

__NTPServers__
List of NTP servers for the DHCP process.

## HTTP JSON API

Here are some examples on how to exercise the JSON API. One day this will be served through a Swagger interface (todo).

Authentication is done via HTTP Basic Auth. Please remember to use `/save` after each API set of POST calls to `/host`.

If you require SSL, please raise an issue on [GitHub](https://github.com/networkbootstrap/ztpmanagerassets.git) and let's chat about it!

__Create Hosts__

Please note, if you want to add an image for Junos to download and update, add the field `imagefile` with the full name of the image in the `/images` directory.

```bash
curl -X POST \
  -H 'Content-Type: application/json' \
  -H "Authorization: Basic YWRtaW46UGFzc3cwcmQ=" \
  -d '{
    "ethernetaddress": "00:0c:29:4d:3d:cd",
    "fixedipaddress": "192.168.50.101",
    "hostname": "demo02",
    "vendor": "junos"
    }' \
   REPLACE_WITH_SERVER_IP:1323/hosts
```

__Delete Hosts__

```bash
curl -X DELETE -H 'Content-Type: application/json' \
    -H "Authorization: Basic YWRtaW46UGFzc3cwcmQ=" \
    REPLACE_WITH_SERVER_IP:1323/hosts/REPLACE_WITH_HOST_IP
```

__Get Hosts__

```bash
curl -X GET -H "Authorization: Basic YWRtaW46UGFzc3cwcmQ=" \
    http://REPLACE_WITH_SERVER_IP:1323/hosts
```

__Get Individual Hosts__

```bash
curl -X GET -H "Authorization: Basic YWRtaW46UGFzc3cwcmQ=" \
    http://REPLACE_WITH_SERVER_IP:1323/hosts/REPLACE_WITH_HOST_IP
```

__Save__

```bash
curl -X POST -H 'Content-Type: application/json' \
    -H "Authorization: Basic YWRtaW46UGFzc3cwcmQ=" \
    REPLACE_WITH_SERVER_IP:1323/save
```

*These docs are work in progress and will be updated regularly*

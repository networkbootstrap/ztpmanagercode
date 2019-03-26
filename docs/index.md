# Welcome

This set of docs will guide you through installation and use of ZTPManager.

## What is this ZTPManager?

ZTPManager is an application manager for the [ISC-DHCP-SERVER](https://www.isc.org/downloads/dhcp/) which configures it for use in Zero Touch Provisioning (ZTP) scenarios. Initially created for Junos but with some feature enhancements can probably be used for other vendors.

__ZTPManager does:__

- Configures `dhcpd.conf` file with basic info and hosts for each ZTP client.
- Creates initial configurations from templates for Juniper devices.
- Manages `isc-dhcp-server` file and serving interfaces.
- Provides a HTTP JSON API for configuration (super simple).
- Installs the ISC-DHCP-SERVER.

You basically get your life back and stop trying to figure out how to configure a DHCP server for ZTP.

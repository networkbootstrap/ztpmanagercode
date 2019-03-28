# Welcome

This set of docs will guide you through installation and use of ZTPManager.

<div style="position: relative; padding-bottom: 56.25%; height: 0; overflow: hidden; max-width: 100%; height: auto;">
    <iframe src="https://www.youtube.com/embed/3Wz4COk-ae4" frameborder="0" allowfullscreen style="position: absolute; top: 0; left: 0; width: 100%; height: 100%;"></iframe>
</div>


## What is this ZTPManager?

ZTPManager is an application manager for the [ISC-DHCP-SERVER](https://www.isc.org/downloads/dhcp/) which configures it for use in Zero Touch Provisioning (ZTP) scenarios. Initially created for Junos but with some feature enhancements can probably be used for other vendors.

__ZTPManager does:__

- Configures `dhcpd.conf` file with basic info and hosts for each ZTP client.
- Creates initial configurations from templates for Juniper devices.
- Manages `isc-dhcp-server` file and serving interfaces.
- Provides a HTTP JSON API for configuration (super simple).
- Installs the ISC-DHCP-SERVER.

You basically get your life back and stop trying to figure out how to configure a DHCP server for ZTP.

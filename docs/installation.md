# Installation Guide

ZTPManager can be installed through a one line install procedure, or you can compile from source and place the files manually.

### One-Line Installer

This is the easiest to get started with ZTPManager but it does rely on pulling a pre-compiled binary down from GitHub for x64_x86 architecture. If you don't like that and you think there might be some nasties hiding away, that's fine. Go check out the manual approach!

The installer does this:

- Install ISC-DHCP-SERVER
- Install git
- Clones the `ztpmanagerassets` repository using git
- Writes elements of the configuration file from script input
- Creates `config` and `images` directories
- Launches `ZTPmanager` with the HTTP file interface running on port 80 and the configuration API running on port 1323.

How to run this wonder?

You must know what the interface on the server is that you would like to serve DHCP. In the example below it's `ens34`.
Secondly, you must know what URL you would like to access the server on. This could be a FQDN or just an IP address.

Make sure you have `curl` installed. On Ubuntu this can be installed with `sudo apt install curl`.

Here is the one line install. Copy and paste this in to your terminal. 

`sudo curl -sSL https://raw.githubusercontent.com/networkbootstrap/ztpmanager/master/install.sh | bash -s -- --url=ztp.simpledemo.net --iface=ens34`

### Manual 

In order to do a manual install, you're required to have the Go tool-chain installed (written in 1.10), install the dependencies, build and them download the required assets to finish off the installation. As the 'New Kids on the Block' would say, step-by-step:

1.	Clone the code repository
`git clone https://github.com/networkbootstrap/ztpmanagercode.git`

2.	Grab the dependencies
`dep ensure`

3.	Compile for your target operating system
`GOOS=linux go build -o ztpmanager`

4.	Create a directory on the target machine
`mkdir ztpmanager`

5.	Copy the binary to the target machine using SCP or your favourite tool

6.	Create `configs` and `images` directory on the target machine
`mkdir configs && mkdir images`

7.	Check that port TCP 80 is free on the target machine, this is where the files are served from. Also check for 1323 or the API port.
`ss -aln | grep "80 "`
`ss -aln | grep "1323 "`

8.	Change config file lines that need changing (see section 'Config File')

9.	Install ISC-DHCP-SERVER version 4.3.5 
` sudo apt-get install -y isc-dhcp-server=4.3.5-3ubuntu7`

10.	Start the binary
`sudo ./binary_name -config config.toml &`

11.	Do a quick check to make sure everything is operational. Read __Usage__ on how to use it.
`curl -X GET REPLACE_WITH_SERVER_IP:1323/hosts`

If you don't have any hosts configured in the `config.toml` then you won't get any meaningful data, but you'll still get a HTTP response code in the 200 range.

*These docs are work in progress and will be updated regularly*

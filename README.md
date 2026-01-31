<!-- [![CircleCI](https://circleci.com/gh/broderickhyman/albiondata-client/tree/master.svg?style=svg)](https://circleci.com/gh/broderickhyman/albiondata-client/tree/master) [![Go Report Card](https://goreportcard.com/badge/github.com/broderickhyman/albiondata-client)](https://goreportcard.com/report/github.com/broderickhyman/albiondata-client)
-->

# Albion Data - Client
Distributed client for the [Albion Online Data](https://www.albion-online-data.com/)
project.

A quick note on the legality of this application and if it
violates the Terms and Conditions for Albion Online. Here is
the response from SBI when asked if we are allowed to do
monitor network packets relating to Albion Online:
> Our position is quite simple. As long as you just look and
analyze we are ok with it. The moment you modify or manipulate
something or somehow interfere with our services we will react
(e.g. perma-ban, take legal action, whatever).

~ MadDave - Technical Lead for Albion Online

Source: https://forum.albiononline.com/index.php/Thread/51604-Is-it-allowed-to-scan-your-internet-trafic-and-pick-up-logs/?postID=512670#post512670

This client monitors local network traffic, identifies UDP packets
that contain relevant data for Albion Online, and ships the information
off to a central NATS server that anyone can subscribe to.

<!--
[Client download stats](https://www.somsubhra.com/github-release-stats/?username=broderickhyman&repository=albiondata-client)
-->

<!-- 
### Contributing
This process is run on a [DigitalOcean Droplet](https://www.digitalocean.com) in order to ensure almost perfect uptime and high performance for the users. If you find this project beneficial to you then please consider a donation, thanks!!

-->

# Contributions
Many thanks to the original developers:
- [Regner](https://github.com/Regner)
- [pcdummy](https://github.com/pcdummy)
- [Ultraporing](https://github.com/Ultraporing)


Many thanks also to [broderickhyman](https://github.com/broderickhyman) for picking up development and funding for the the last few years of the project!

As of 2023-01-01, [Stanx](https://github.com/phendryx) is the primary maintainer and provides funding of the related projects.  

[Walkynn](https://github.com/walkeralencar) has been a long time maintainer of different aspets of the project as well.

# Downloads
Downloads can be found here: https://github.com/ao-data/albiondata-client/releases

Stats for the client releases can be viewed [here](https://tooomm.github.io/github-release-stats/?username=ao-data&repository=albiondata-client).
## Running on Mac

### Running from the Finder
1. Download the latest `albiondata-client-amd64-mac.zip` file from [the Releases page](https://github.com/ao-data/albiondata-client/releases)
2. Unzip that file from the Finder
3. Enter the `albiondata-client` folder.
4. Double click the `run.command` file. It will ask for your password for permissions reasons.

### Running from the Terminal
1. Download the latest `update-darwin-amd64.gz` file from [the Releases page](https://github.com/ao-data/albiondata-client/releases)
2. Unzip that file from the Finder or with `gunzip update-darwin-amd64.gz`
3. The unzipped `albiondata-client` file is a Golang binary file. You'll need to make this file executable so it can be run directly. You can do this from your Terminal with: `chmod +x albiondata-client`
4. Run the client from your Terminal with `./albiondata-client`

## Running on Debian or Debian based distros

### Install app binary
<sup>`~/.local/bin` requires systemd. If you don't roll with systemd use something else. </sup>

1. Create ~/.local/bin folder: `mkdir -p ~/.local/bin`
2. Download latest `update-linux-amd64.gz` version from [the Releases page](https://github.com/ao-data/albiondata-client/releases)  
`curl -L https://github.com/ao-data/albiondata-client/releases/latest/download/update-linux-amd64.gz -o - | gzip -d > ~/.local/bin/albiondata-client`
3. Give user execution permission: `chmod u+x ~/.local/bin/albiondata-client`

### Install dependency libpcap

```bash
sudo apt install libpcap-dev
```

### Give binary permission to capture network traffic

To allow binary to capture data without using sudo

```bash
sudo setcap cap_net_raw,cap_net_admin=eip ~/.local/bin/albiondata-client
```

# Related Projects
- [albiondata-deduper-dotNet](https://github.com/ao-data/albiondata-deduper-dotNet)
- [albiondata-sql-dotNet](https://github.com/ao-data/albiondata-sql-dotNet)
- [albiondata-api-dotNet](https://github.com/ao-data/albiondata-api-dotNet)
- [AlbionData.Models](https://github.com/ao-data/albiondata-models-dotNet) [![NuGet](https://img.shields.io/nuget/v/AlbionData.Models.svg)](https://www.nuget.org/packages/AlbionData.Models/)
- [albion-data-website](https://github.com/ao-data/albion-data-website)

# Contact Us
The best way to get in touch with us is on the Albion Online Fansites Discord server in either the #proj-albiondata or the #developers channel. A permanent invite link can be found here: [https://discord.gg/TjWdq24](https://discord.gg/TjWdq24)

# Developer Setup
### Mac/Linux Setup
- Install go
- Build the project (Go modules will download automatically)

### Windows Setup
[Windows Setup Guide](https://github.com/ao-data/albiondata-client/wiki/Building-in-Windows)

# License
This project, and all contributed code, are licensed under the MIT
License. A copy of the MIT License may be found in the repository.

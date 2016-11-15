# ifconfigdockerhost

This is a pseudo daemon for osx only, that relyable calls "ifconfig alias lo0 192.168.23.1" for local development purposes.

## Motivation

In development we are using the IP address **192.168.23.1** to have a relyable IP Address for **dockerhost** in changing network environments. To achieve this, we call ```sudo ifconfig lo0 alias 192.168.23.1```. Unfourtunately there is no good way to do this ... therefor this __daemon__.

## How does it work ?

It installs a launch daemon to /Library/LaunchDaemons/org.foomo.ifconfigdockerhost.plist and this program to /usr/local/bin. It also does initial loading etc for you.

## Usage

Installation

```bash
go get -u github.com/foomo/ifconfigdockerhost
sudo $GOPATH/bin/ifconfigdockerhost -install
```

- use launchctl for instaraction
- logs go to /var/log/system.log

That´s it.






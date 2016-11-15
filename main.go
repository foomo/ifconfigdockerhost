package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"log/syslog"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const plist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>

	<key>Label</key>
	<string>org.foomo.ifconfigdockerhost</string>

	<key>Program</key>
	<string>/usr/local/bin/ifconfigdockerhost</string>

	<key>RunAtLoad</key>
	<true/>
	
</dict>
</plist>`

const name = "org.foomo.ifconfigdockerhost"

func install() error {

	// copy binary
	binaryName := os.Args[0]
	binaryBytes, readErr := ioutil.ReadFile(binaryName)
	if readErr != nil {
		fmt.Println("did you call me with my full path ?")
		return readErr
	}
	binaryWriteErr := ioutil.WriteFile("/usr/local/bin/ifconfigdockerhost", binaryBytes, 0777)
	if binaryWriteErr != nil {
		return binaryWriteErr
	}

	// write plist
	pListFile := "/Library/LaunchDaemons/" + name + ".plist"
	writeErr := ioutil.WriteFile(pListFile, []byte(plist), 0644)
	if writeErr != nil {
		return writeErr
	}

	// load plist
	launchctl := func(args ...string) error {
		out, loadErr := exec.Command("launchctl", args...).CombinedOutput()
		fmt.Println("	", string(out))
		if loadErr != nil {
			return loadErr
		}
		return nil
	}
	fmt.Println("stopping")
	launchctl("stop", name)

	fmt.Println("unloading")
	launchctl("unload", pListFile)

	fmt.Println("loading plist")
	loadErr := launchctl("load", pListFile)
	if loadErr != nil {
		return loadErr
	}

	// launch shit
	out, startErr := exec.Command("launchctl", "start", name).CombinedOutput()
	fmt.Println("	", string(out))
	if startErr != nil {
		return startErr
	}
	return nil
}

func ifconfig(ip string, l *syslog.Writer) error {
	cmd := exec.Command("ifconfig", "lo0", "alias", ip)
	combinedOut, runErr := cmd.CombinedOutput()
	combinedOutStr := string(combinedOut)
	if combinedOutStr != "" {
		l.Info(combinedOutStr)
	}
	if runErr != nil {
		return runErr
	}
	return nil
}

func up(ip string) (bool, error) {
	addressses, err := net.InterfaceAddrs()
	if err != nil {
		return false, err
	}
	for _, addr := range addressses {
		//fmt.Println("addr", addr.String(), "network", addr.Network())
		addressParts := strings.Split(addr.String(), "/")
		if len(addressParts) > 0 {
			if addressParts[0] == ip {
				return true, nil
			}
		}
	}
	return false, nil
}

func run(ip string, l *syslog.Writer) {
	l.Info("running for: " + ip)
	upErrCounter := 0
	ifconfigErrorCounter := 0
	for {
		up, upErr := up(ip)
		if upErr == nil {
			upErrCounter = 0
		} else {
			upErrCounter++
			l.Warning("could not determine up status")
			if upErrCounter > 10 {
				l.Err("giving up, too many errors while trying to determine up status: " + ip + " " + upErr.Error())
				os.Exit(1)
			}
		}
		if !up {
			l.Warning("ip not up, trying to ifconfig")
			err := ifconfig(ip, l)
			if err != nil {
				ifconfigErrorCounter++
				l.Warning("could not ifconfig: " + err.Error())
				if ifconfigErrorCounter > 10 {
					l.Err("giving up, too many errors while trying to bring up: " + ip + " " + err.Error())
					os.Exit(1)
				}
			} else {
				ifconfigErrorCounter = 0
			}
		}
		time.Sleep(time.Second * 1)
	}
}

func main() {
	if runtime.GOOS != "darwin" {
		log.Fatalln("yes this is a go program but it is only for osx and not for", runtime.GOOS)
	}
	flagInstall := flag.Bool("install", false, "install the osx daemon")
	flagIP := flag.String("ip", "192.168.23.1", "ip to add to local loopback interface lo0")
	flag.Parse()
	if *flagInstall {
		err := install()
		if err != nil {
			fmt.Println("install failed", err)
			os.Exit(1)
		}
		fmt.Println("ok, you are ready to go")
		os.Exit(0)
	}
	l, le := syslog.New(syslog.LOG_DAEMON, "org.foomo.ifconfigdockerhost")
	if le != nil {
		log.Fatal("failed to get system log writer: " + le.Error())
	}
	run(*flagIP, l)
}

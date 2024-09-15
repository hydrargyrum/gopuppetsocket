// SPDX-License-Identifier: WTFPL

package main

import "flag"
import "fmt"
import "log"
import "time"
import "os"
import "net"

var realAddress = flag.String("real", "", "ADDRESS:PORT of the real end server that should be reached")
var puppetAddress = flag.String("puppet", "", "ADDRESS:PORT of the puppet server")
var canWaitLonger = flag.Bool("can-wait-longer", false, "wait 1 minute between connects if no puppet connection for 24 hours")
var checkAddrLater = flag.Bool("check-addr-later", false, "don't check addresses at start (start client even if network is down)")

func copyTo(from, to net.Conn) {
	buf := make([]byte, 1024)

	for {
		nread, err := from.Read(buf)
		if err != nil {
			log.Printf("error when reading: %s", err)
			to.Close()
			return
		}

		subbuf := buf[:nread]

		for len(subbuf) > 0 {
			nwrite, err := to.Write(subbuf)
			if err != nil {
				log.Printf("error when writing: %s", err)
				from.Close()
				return
			}

			subbuf = subbuf[nwrite:]
		}
	}
}

func checkAddr(addr string, label string) {
	if len(addr) == 0 {
		fmt.Fprintf(flag.CommandLine.Output(), "%s address cannot be empty\n", label)
		flag.Usage()
		os.Exit(1)
	}

	if (*checkAddrLater) {
		return
	}
	if _, err := net.ResolveTCPAddr("tcp", addr); err != nil {
		log.Fatalf("could not resolve %s address: %s", label, err)
	}
}

var logEveryNs = time.Duration(5 * 1_000_000_000)
var connectWaitNs = time.Duration(1_000_000_000)
var connectLongWaitNs = time.Duration(60 * 1_000_000_000)
var delayAfterNs = time.Duration(24 * 60 * 60 * 1_000_000_000)

func main() {
	flag.Parse()

	checkAddr(*realAddress, "real server")
	checkAddr(*puppetAddress, "puppet server")

	ticker := time.Now()

	lastConnected := time.Now()

	for {
		puppetServer, err := net.Dial("tcp", *puppetAddress)
		if err != nil {
			if time.Since(ticker) > logEveryNs {
				log.Printf("could not connect to puppet server: %s", err)
				ticker = time.Now()
			}

			if *canWaitLonger && time.Since(lastConnected) >= delayAfterNs {
				time.Sleep(connectLongWaitNs)
			} else {
				time.Sleep(connectWaitNs)
			}

			continue
		}
		log.Printf("connected to puppet server %v", puppetServer.RemoteAddr())

		lastConnected = time.Now()

		realServer, err := net.Dial("tcp", *realAddress)
		if err != nil {
			log.Printf("could not connect to real server: %s", err)
			puppetServer.Close()
			continue
		}
		log.Printf("connected to real server %v, bridging to puppet server %v", realServer.RemoteAddr(), puppetServer.RemoteAddr())

		go copyTo(puppetServer, realServer)
		go copyTo(realServer, puppetServer)
	}
}

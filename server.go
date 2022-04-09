package main

import "flag"
import "fmt"
import "log"

//import "time"
import "os"
import "net"
import "strconv"

var realAddress = flag.String("real", "", "ADDRESS:PORT on which to listen connections from a real client")
var puppetAddress = flag.String("puppet", "", "ADDRESS:PORT on which to listen for the puppet client")

func copyTo(from, to net.Conn) {
	buf := make([]byte, 1024)

	for {
		nread, err := from.Read(buf)
		if err != nil {
			log.Printf("error when reading: %s", err)
			to.Close()
			return
		}

		log.Printf("read %d bytes", nread)

		subbuf := buf[:nread]

		for len(subbuf) > 0 {
			nwrite, err := to.Write(subbuf)
			if err != nil {
				log.Printf("error when writing: %s", err)
				from.Close()
				return
			}

			log.Printf("wrote %d bytes", nwrite)
			subbuf = subbuf[nwrite:]
		}
	}
}

func handleRealClient(realClient net.Conn) {
	defer realClient.Close()

	puppetListener, err := net.Listen("tcp", *puppetAddress)
	if err != nil {
		log.Printf("could not listen for puppet clients: %s", err)
		return
	}

	puppetClient, err := puppetListener.Accept()
	puppetListener.Close()
	if err != nil {
		log.Printf("could not accept a puppet connection: %s", err)
		return
	}
	log.Printf("accepted puppet client %s", puppetClient.RemoteAddr().String())

	go copyTo(puppetClient, realClient)
	copyTo(realClient, puppetClient)
}

func checkAddr(addr string, label string) string {
	if len(addr) == 0 {
		fmt.Fprintf(flag.CommandLine.Output(), "%s address cannot be empty\n", label)
		flag.Usage()
		os.Exit(1)
	}

	_, _, err := net.SplitHostPort(addr)
	switch err := err.(type) {
	case nil:
	case *net.AddrError:
		if err.Err == "missing port in address" {
			_, err := strconv.Atoi(addr)
			if err != nil {
				log.Fatalf("%s address is not in ADDRESS:PORT or PORT format", label)
			}
			addr = fmt.Sprintf("0.0.0.0:%s", addr)
		}
	default:
		log.Fatalf("could not parse %s address: %s", label, err)
	}

	if _, err := net.ResolveTCPAddr("tcp", addr); err != nil {
		log.Fatalf("could not resolve %s address: %s", label, err)
	}

	return addr
}

func main() {
	flag.Parse()

	*realAddress = checkAddr(*realAddress, "real")
	*puppetAddress = checkAddr(*puppetAddress, "puppet")

	realListener, err := net.Listen("tcp", *realAddress)
	if err != nil {
		log.Fatalf("could not listen for real clients: %s", err)
	}

	for {
		realClient, err := realListener.Accept()
		if err != nil {
			log.Printf("oops accepting real client %s", err)
			continue
		}
		log.Printf("accepted real client %s", realClient.RemoteAddr().String())

		go handleRealClient(realClient)
	}
}

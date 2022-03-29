package main

import "log"
import "time"
import "os"
import "net"

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

func main() {
	for {
		puppetServer, err := net.Dial("tcp", os.Args[1])
		if err != nil {
			log.Printf("could not connect to puppet server: %s", err)
			time.Sleep(time.Duration(1000 * 1000 * 1000))
			continue
		}
		log.Printf("connected to puppet server")

		realServer, err := net.Dial("tcp", os.Args[2])
		if err != nil {
			log.Printf("could not connect to real server: %s", err)
			puppetServer.Close()
			continue
		}
		log.Printf("connected to real client")

		go copyTo(puppetServer, realServer)
		go copyTo(realServer, puppetServer)
	}
}

package main

import "log"

//import "time"
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

func handleRealClient(realClient net.Conn) {
	defer realClient.Close()

	puppet_listener, err := net.Listen("tcp", os.Args[2])
	if err != nil {
		log.Printf("could not listen for puppet clients: %s", err)
		return
	}

	puppetClient, err := puppet_listener.Accept()
	puppet_listener.Close()
	if err != nil {
		log.Printf("could not accept a puppet connection: %s", err)
		return
	}
	log.Printf("accepted puppet client %s", puppetClient)

	go copyTo(puppetClient, realClient)
	copyTo(realClient, puppetClient)
}

func main() {
	realListener, err := net.Listen("tcp", os.Args[1])
	if err != nil {
		log.Fatalf("could not listen for real clients: %s", err)
	}

	for {
		realClient, err := realListener.Accept()
		if err != nil {
			log.Printf("oops accepting real client %s", err)
			continue
		}
		log.Printf("accepted real client %s", realClient)

		go handleRealClient(realClient)
	}

}

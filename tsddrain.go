package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

const (
	versionString = "tsddrain v1.0\n"
)

var (
	tHost         = flag.String("b", "localhost:4242", "bind host")
	outFile       = flag.String("o", "", "output file")
	flushInterval = flag.Duration("f", time.Second*2, "flush interval")
)

type ServerStat struct {
	events  int64
	bytes   int64
	unknown int64
	conns   int
	open    int
}

func main() {

	flag.Parse()
	stats := ServerStat{}

	msgChan := make(chan string, 16)
	if *outFile != "" && *outFile != "-" {
		go writeFile(msgChan)
	}

	s, err := net.Listen("tcp", *tHost)
	if err != nil {
		log.Fatalln("Couldn't start server", err)
	}

	for {
		conn, err := s.Accept()
		if err != nil {
			log.Println("Failed to accept connection", err)
			continue
		}
		log.Println("Accepted new connection from", conn.RemoteAddr().String())

		go func() {
			defer func() {
				stats.open -= 1
				conn.Close()
			}()
			stats.conns++
			stats.open++

			scanner := bufio.NewScanner(conn)
			writer := bufio.NewWriter(conn)

			for scanner.Scan() {
				// extending the deadline on every loop pass seems to be idiomatic, but expensive
				//conn.SetReadDeadline(time.Now().Add(time.Duration(6000 * time.Millisecond)))
				event := scanner.Text()
				switch {
				case strings.HasPrefix(event, "put "):
					stats.events++
					stats.bytes += int64(len(event))
					if *outFile != "" {
						if *outFile == "-" {
							log.Println(event)
						} else {
							msgChan <- event
						}
					}
				case strings.HasPrefix(event, "version"):
					writer.WriteString(versionString)
					writer.Flush()
				case strings.HasPrefix(event, "stats"):
					t := time.Now().Unix()
					fmt.Fprintf(writer, "tsdsink.events %d %d\n", t, stats.events)
					fmt.Fprintf(writer, "tsdsink.bytes %d %d\n", t, stats.bytes)
					fmt.Fprintf(writer, "tsdsink.connections %d %d\n", t, stats.conns)
					fmt.Fprintf(writer, "tsdsink.open %d %d\n", t, stats.open)
					fmt.Fprintf(writer, "tsdsink.unknown %d %d\n", t, stats.unknown)
					writer.Flush()
				case strings.HasPrefix(event, "reset"):
					log.Println("Resetting stats")
					stats = ServerStat{}
				default:
					log.Println("Unknown command", event)
					stats.unknown++
				}
			}
			if err := scanner.Err(); err != nil {
				log.Println("Error reading standard input", err)
			}

		}()
	}

	log.Println("Listening on %s for connections", tHost)

	select {}
}

func writeFile(msgChan chan string) {
	f, err := os.OpenFile(*outFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		log.Fatalln("Error opening file:", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	ticker := time.Tick(*flushInterval)

	for {
		select {
		case event := <-msgChan:
			_, err = fmt.Fprintln(w, event)
			if err != nil {
				log.Println("Error writing to file", err)
			}
		case <-ticker:
			w.Flush()
		}
	}
}

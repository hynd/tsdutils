package main

import (
    "bufio"
    "net"
    "fmt"
    "os"
    "flag"
    "strings"
    "time"
)

const (
  versionString = "tsddrain v1.0\n"
)

var (
  tHost         = flag.String("b", "localhost:4242", "bind host")
  outFile       = flag.String("o", "/tmp/tsd.out", "output file")
  flushInterval = flag.Duration("f", time.Second * 2, "flush interval")
)

func main() {

  flag.Parse()
  msgChan := make(chan string, 16)

  go func() {
    f, err := os.OpenFile(*outFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
    if err != nil {
      fmt.Println("Error opening file", *outFile, err)
      os.Exit(1)
    }
    defer f.Close()

    w := bufio.NewWriter(f)
    ticker := time.Tick(*flushInterval)

    for {
      select {
        case event := <-msgChan:
          _, err = fmt.Fprintln(w, event)
          if err != nil {
            fmt.Println("Error writing to file", err)
          }
        case <-ticker:
          w.Flush()
      }
    }
  }()


  s, err := net.Listen("tcp", *tHost)
  if err != nil {
    fmt.Println("Couldn't start server", err)
    os.Exit(1)
  }

  for {
    conn, err := s.Accept()
    if err != nil {
      fmt.Println("Failed to accept connection", err)
      continue
    }
    fmt.Println("Accepted new connection from", conn.RemoteAddr().String())

    go func() {
      defer conn.Close()

      scanner := bufio.NewScanner(conn)
      writer  := bufio.NewWriter(conn)

      for scanner.Scan() {
        event := scanner.Text()
        switch {
          case strings.HasPrefix(event, "put "):
            msgChan<- event
          case strings.HasPrefix(event, "version"):
            writer.WriteString(versionString)
            writer.Flush()
          default:
            fmt.Println("Unknown command", event)
        }
      }
      if err := scanner.Err(); err != nil {
        fmt.Println("Error reading standard input", err)
      }

    }()
  }

  fmt.Println("Listening on %s for connections", tHost)

  select{}
}


package main

import (
	"bytes"
	"context"
	"log"
	"net"
	"time"
)

func main() {

	var dialer net.Dialer

	ctx, _ := context.WithTimeout(context.Background(), time.Second*1)
	conn, err := dialer.DialContext(ctx, "tcp", ":5005")
	if err != nil {
		log.Println("failed to dial", err)
		return
	}
	var buf1 bytes.Buffer
	buf1.WriteString("hello")

	n, err := buf1.WriteTo(conn)
	if err != nil {
		log.Println("failed to write", err)
		return
	}
	log.Println("wrote bytes", n)

	var buf bytes.Buffer

	// _, err = io.Copy(&buf, conn) // only works if the server don't care for the response and sends EOF
	// _, err = bufio.NewReader(conn).Read(buf.Bytes())
	n, err = buf.ReadFrom(conn)
	if err != nil {
		log.Println("failed to read to buff", err)
		return
	}

	log.Println("got", buf.String(), n)
}

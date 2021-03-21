package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"
)

func main() {

	var dialer net.Dialer

	var ctx, cancel = context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	conn, err := dialer.DialContext(ctx, "udp", ":5005")
	if err != nil {
		log.Fatalln("failed to dial", err)
	}
	defer conn.Close()

	// otherwise read from conn

	var resp = []byte("hello")
	_, err = conn.Write(resp)
	if err != nil {
		log.Fatalln("failed to write", err)
	}

	var recvd = make([]byte, 128)
	_, err = conn.Read(recvd)
	if err != nil {
		log.Fatalln("failed to read", err)
	}

	fmt.Println("got", string(recvd))
}

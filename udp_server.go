package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// This is an echo sever on top of raw udp protocol with
// * graceful handling of interrupt and cancellation signals
// * avoids "use of closed network connection" error in blocked call to conn.ReadFrom() at termination of the server
func main() {
	var ctx, cancel = context.WithCancel(context.Background())

	go func() {
		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		fmt.Println("cancelling")
		cancel()
	}()
	var doneChan = make(chan struct{})

	pConn, err := net.ListenPacket("udp", ":5005")
	if err != nil {
		log.Fatalln("failed to listen", err)
	}
	// we launch the blocking conn.ReadFrom in a separate goroutine
	go handleConnection(ctx, pConn, doneChan)

	// block until the handling loop is finished
	select {
	case <-doneChan:
		log.Println("finished")
	}

}

type readTimeout error

var ReadTimeout readTimeout = errors.New("read timeout on waiting udp connection")

func handleConnection(ctx context.Context, conn net.PacketConn, finishChan chan struct{}) {

	// main loop for continously handling each packet
	// since there is no concept of a connection with udp
	// we handle each packet in a separate goroutine
	for {

		buffer := make([]byte, 4096)
		// we mimic a polling behaviour on conn.ReadFrom by extending the read deadline on the packetConn continously in a loop
	loop:
		select {
		// if we got call to terminate we are done
		case <-ctx.Done():
			// signal finish to the main fn
			finishChan <- struct{}{}
			fmt.Println("exiting loop")
			return
		default:
		}
		err := conn.SetReadDeadline(time.Now().Add(time.Second * 5))
		if err != nil {
			log.Println("failed to set read deadline", err)
		}
		// idle until we get a packet or time out
		n, raddr, err := conn.ReadFrom(buffer)

		if err != nil {
			opErr, ok := err.(*net.OpError)
			if ok && opErr.Timeout() {
				// read timeout most likely;
				// so we should loop again
				goto loop
			}
			log.Println("failed to read from conn", err)
			goto loop
		}

		// we got a packet, write back whatever is read
		_, err = conn.WriteTo(buffer[:n], raddr)
		if err != nil {
			log.Println("failed to write to conn", err)
		}

	}

}

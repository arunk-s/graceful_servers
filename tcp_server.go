package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// This is an echo sever on top of raw tcp protocol with
// * graceful handling of interrupt and cancellation signals
// * avoids "use of closed network connection" error in blocked call to listener.Accept() at termination of the server
func main() {

	addr, err := net.ResolveTCPAddr("tcp", ":5005")
	if err != nil {
		log.Println("failed to resolve", err)
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Println("failed to listen", err)
	}

	// doneChan used to signal connection handler loop exit
	var doneChan = make(chan struct{})
	var ctx, cancel = context.WithCancel(context.Background())

	var wg sync.WaitGroup
	// register for SIGINT and SIGTERM signal
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		fmt.Println("cancelling")
		cancel()
	}()

	// main loop for handing off each recvd connection to a goroutine
	// we don't just do l.Accept() and hand off to avoid  "connection is closed"
	go func() {
	loop:
		// Why are we doing this ?
		// Simply to catch the cancellation signal but not blocked at l.Accept()
		// calling l.Close() while we are blocked on l.Accept() will produce error "use of closed network connection"
		// so we mimic a polling loop on l.Accept() with the help of SetDeadline()
		// If we get a done signal via ctx, wait for existing connection to finish then close the listener and don't go accepting again
		// else extend deadline again and continue polling
		for {
		again:
			select {
			case <-ctx.Done():
				// wait for live connections to finish
				wg.Wait()
				l.Close()
				fmt.Println("exiting loop")
				// send signal to the main fn that we are done
				doneChan <- struct{}{}
				break loop
			default:
			}
			l.SetDeadline(time.Now().Add(time.Second * 2))
			conn, err := l.Accept()
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				// extend deadline again
				goto again
			} else if err != nil {
				log.Println("failed to listen", err)
			}
			wg.Add(1)
			go handleConnection(conn, &wg)
		}
	}()

	// block forever until the connection loop exits
	select {
	case <-doneChan:
		// the connection loop exits
		fmt.Println("exiting")
	}
}

func handleConnection(conn net.Conn, wg *sync.WaitGroup) {

	defer conn.Close()
	// finish wg when the conn is read
	defer wg.Done()

	// var buffer bytes.Buffer
	// buffer.Grow(4096)
	// n, err := io.Copy(&buffer, conn) // this will only work if the client sends an EOF after each packet that fits into 4096 bytes and don't wait for the response of the server on the same connection
	// n, err := bufio.NewReader(conn).Read(buffer.Bytes())
	// n, err := buffer.ReadFrom(conn) // can't use read from because it also reads until EOF

	buffer := make([]byte, 128)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Println("failed to read from conn", err)
		return
	}
	// log.Println("read from conn", buffer.String(), n)

	// write back to client
	n2, err := conn.Write(buffer[:n])
	// n2, err := conn.Write(buffer.Bytes())
	if err != nil {
		log.Println("failed to write conn", err)
		return
	}
	log.Println("wrote to conn", n2)

}

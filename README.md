## TCP and UDP echo servers written in Go 

The echo servers have following features

* graceful termination on SIGINT/interrupt and SIGTERM/terminate signal
* avoids "use of closed network connection" errors when terminating the server

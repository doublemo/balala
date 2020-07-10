package test

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/doublemo/balala/cores/networks"
)

var usageStr = `
Usage: balala [options]
Server Options:
    -a, --addr <host>                Bind to host address (default: 0.0.0.0)
    -p, --port <port>                Use port for clients (default: 4222)
    -P, --pid <file>                 File to store PID
    -m, --http_port <port>           Use port for http monitoring
    -ms,--https_port <port>          Use port for https monitoring
    -c, --config <file>              Configuration file
    -sl,--signal <signal>[=<pid>]    Send signal to nats-server process (stop, quit, reopen, reload)
                                     <pid> can be either a PID (e.g. 1) or the path to a PID file (e.g. /var/run/server.pid)
        --client_advertise <string>  Client URL to advertise to other servers
    -t                               Test configuration and exit
Logging Options:
    -l, --log <file>                 File to redirect log output
    -T, --logtime                    Timestamp log entries (default: true)
    -s, --syslog                     Log to syslog or windows event log
    -r, --remote_syslog <addr>       Syslog server addr (udp://localhost:514)
    -D, --debug                      Enable debugging output
    -V, --trace                      Trace the raw protocol
    -DV                              Debug and trace
Authorization Options:
        --user <user>                User required for connections
        --pass <password>            Password required for connections
        --auth <token>               Authorization token required for connections
TLS Options:
        --tls                        Enable TLS, do not verify clients (default: false)
        --tlscert <file>             Server certificate file
        --tlskey <file>              Private key for server certificate
        --tlsverify                  Enable TLS, verify client certificates
        --tlscacert <file>           Client certificate CA for verification
Cluster Options:
        --routes <rurl-1, rurl-2>    Routes to solicit and connect
        --cluster <cluster-url>      Cluster URL for solicited routes
        --no_advertise <bool>        Advertise known cluster IPs to clients
        --cluster_advertise <string> Cluster URL to advertise to other servers
        --connect_retries <number>   For implicit routes, number of connect retries
Common Options:
    -h, --help                       Show this message
    -v, --version                    Show version
        --help_tls                   TLS help
`

func usage() {
	fmt.Printf("%s\n", usageStr)
	os.Exit(0)
}

func main() {
	var (
		showHelp bool
	)

	fs := flag.NewFlagSet("test", flag.ExitOnError)
	fs.Usage = usage

	args := os.Args[1:]

	fs.BoolVar(&showHelp, "h", false, "Show this message.")
	fs.BoolVar(&showHelp, "help", false, "Show this message.")
	if err := fs.Parse(args); err != nil {
		log.Fatal(err)
		return
	}

	if showHelp {
		fmt.Println("fmt.Println(os.Args)")
	}

	s := networks.NewKCP()
	s.CallBack(func(conn net.Conn, exit chan struct{}) {
		defer func() {
			log.Println("offline:", conn.RemoteAddr())
		}()

		log.Println("from:", conn.RemoteAddr())
		for {
			select {
			case <-exit:
				return
			}
		}
	})

	go func() {
		time.Sleep(20 * time.Second)
		s.Shutdown()
	}()

	go func() {
		time.Sleep(20 * time.Second)
		s.Shutdown()
	}()

	go func() {
		defer func() {
			log.Println("server closed")
		}()
		log.Println(s.Serve(networks.NewKCPDefaultConfig()))
	}()

	select {}
}

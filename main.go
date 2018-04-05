package main

import (
	"flag"
	"fmt"
	"os"
	"net"
	"github.com/gostones/huskie/tunnel"
	"github.com/gostones/huskie/chat"
	"github.com/gostones/huskie/ssh"
	"bytes"
	"strings"
	"time"
	"github.com/gostones/huskie/rp"
	"github.com/gostones/huskie/docker"
)

//
var help = `
  Usage: chisel [command] [--help]

  Commands:
    server - runs chisel in server mode
    client - runs chisel in client mode

  Read more:
    https://github.com/jpillora/chisel

`
//
var rps = `
[common]
bind_port = %v
`
//
var rpc = `
[common]
server_addr = 127.0.0.1
server_port = %v

[ssh]
type = tcp
local_ip = 127.0.0.1
local_port = %v
remote_port = %v
`

func main() {

	flag.Bool("help", false, "")
	flag.Bool("h", false, "")
	flag.Usage = func() {}
	flag.Parse()

	//
	args := flag.Args()

	subcmd := ""
	if len(args) > 0 {
		subcmd = args[0]
		args = args[1:]
	}

	huskiePort := os.Getenv("HUSKIE_PORT")

	serverPort := 7000
	remotePort := 6000

 	//
	switch subcmd {
	case "harness":
		go chat.Server([]string{"--bind", ":" + huskiePort, "--identity", os.Getenv("HUSKIE_IDENTITY")})
		go rp.Server(fmt.Sprintf(rps, serverPort))
		
		tunnel.TunServer(args)
	case "mush":
		port := freePort()
		fmt.Fprintf(os.Stdout, "local: %v remote: %v\n", port, remotePort)

		go tunnel.TunClient(append(args, fmt.Sprintf("localhost:%v:localhost:%v", port, remotePort)))
		//
		for i := 0;;i++ {
			rc := ssh.Client([]string{"--p", fmt.Sprintf("%v", port), "--i", os.Getenv("HUSKIE_IDENTITY"), "ubuntu@localhost"})
			if rc == 0 {
				os.Exit(0)
			}
			secs := time.Duration((i % 10) * (i % 10)) * time.Second
			fmt.Fprintf(os.Stdout, "rc: %v sleeping %v\n", rc, secs)

			time.Sleep(secs)
		}
	case "dog":
		port := freePort()
		servicePort := freePort()
		fmt.Fprintf(os.Stdout, "local: %v remote: %v service: %v\n", port, remotePort, servicePort)

		go docker.Server([]string{"ssh2docker", "--bind", fmt.Sprintf(":%v", servicePort)})
		go tunnel.TunClient(append(args, fmt.Sprintf("localhost:%v:localhost:%v", port, serverPort)))

		for i := 0;;i++ {
			rc := rp.Client(fmt.Sprintf(rpc, port, servicePort, remotePort))
			secs := time.Duration((i % 10) * (i % 10)) * time.Second
			fmt.Fprintf(os.Stdout, "rc: %v sleeping %v\n", rc, secs)

			time.Sleep(secs)
		}
	case "whistle":
		port := freePort()
		user := fmt.Sprintf("u_%v_%v", strings.Replace(macAddr(), ":", "", -1), port)
		fmt.Fprintf(os.Stdout, "local: %v user: %v\n", port, user)

		go tunnel.TunClient(append(args, fmt.Sprintf("localhost:%v:localhost:%v", port, huskiePort)))
		//
		for i := 0;;i++ {
			rc := ssh.Client([]string{"--p", fmt.Sprintf("%v", port), "--i", os.Getenv("HUSKIE_IDENTITY"), user + "@localhost"})

			secs := time.Duration((i % 10) * (i % 10)) * time.Second
			fmt.Fprintf(os.Stdout, "rc: %v sleeping %v\n", rc, secs)

			time.Sleep(secs)
		}
	default:
		fmt.Fprintf(os.Stderr, help)
		os.Exit(1)
	}
}

func freePort() int {
	l, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		panic(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func macAddr() (addr string) {
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, i := range interfaces {
			if i.Flags&net.FlagUp != 0 && bytes.Compare(i.HardwareAddr, nil) != 0 {
				addr = i.HardwareAddr.String()
				break
			}
		}
	}
	return
}
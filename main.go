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
)

var help = `
  Usage: chisel [command] [--help]

  Commands:
    server - runs chisel in server mode
    client - runs chisel in client mode

  Read more:
    https://github.com/jpillora/chisel

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

	switch subcmd {
	case "server":
		go chat.Server([]string{"--bind", ":" + huskiePort, "--identity", os.Getenv("HUSKIE_IDENTITY")})
		tunnel.TunServer(args)
	case "client":
		freePort := freePort()
		user := fmt.Sprintf("u_%v_%v", strings.Replace(macAddr(), ":", "", -1), freePort)
		fmt.Fprintf(os.Stdout, "port: %v user: %v\n", freePort, user)

		tun := fmt.Sprintf("localhost:%v:localhost:%v", freePort, huskiePort)

		args = append(args, tun)
		fmt.Fprintf(os.Stdout, "args: %v\n", args)

		go tunnel.TunClient(args)
		for i := 0;;i++ {
			rc := ssh.Client([]string{"--p", fmt.Sprintf("%v", freePort), "--i", os.Getenv("HUSKIE_IDENTITY"), user + "@localhost"})

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
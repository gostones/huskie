package main

import (
	"flag"
	"fmt"
	"github.com/gostones/huskie/chat"
	"github.com/gostones/huskie/docker"
	"github.com/gostones/huskie/rp"
	"github.com/gostones/huskie/ssh"
	"github.com/gostones/huskie/tunnel"
	"github.com/gostones/huskie/util"
	"os"
	"strings"
)

//
var help = `
	Usage: huskie [command] [--help]

	Commands:
		harness - server mode
		whistle - chat service
		dog     - worker mode (docker instance)
		mush    - control agent
		sshd    - peer host service (bash shell)
		ssh     - connect to peer host
`

//
var rps = `
[common]
bind_port = %v
`

//server_port, local_port, remote_port
var rpc = `
[common]
server_addr = 127.0.0.1
server_port = %v
http_proxy =

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

	sleep := util.BackoffDuration()

	//
	switch subcmd {
	case "harness":
		go chat.Server([]string{"--bind", ":" + huskiePort, "--identity", os.Getenv("HUSKIE_IDENTITY")})
		go rp.Server(fmt.Sprintf(rps, serverPort))

		tunnel.TunServer(args)
	case "mush":
		port := util.FreePort()
		fmt.Fprintf(os.Stdout, "local: %v remote: %v\n", port, remotePort)

		go tunnel.TunClient(append(args, fmt.Sprintf("localhost:%v:localhost:%v", port, remotePort)))
		//
		for {
			rc := ssh.Client([]string{"--p", fmt.Sprintf("%v", port), "--i", os.Getenv("HUSKIE_IDENTITY"), "ubuntu@localhost"})
			if rc == 0 {
				os.Exit(0)
			}

			sleep(rc)
		}
	case "dog":
		port := util.FreePort()
		servicePort := util.FreePort()
		fmt.Fprintf(os.Stdout, "local: %v remote: %v service: %v\n", port, remotePort, servicePort)

		go docker.Server([]string{"ssh2docker", "--bind", fmt.Sprintf(":%v", servicePort)})
		go tunnel.TunClient(append(args, fmt.Sprintf("localhost:%v:localhost:%v", port, serverPort)))

		for {
			rc := rp.Client(fmt.Sprintf(rpc, port, servicePort, remotePort))
			sleep(rc)
		}
	case "whistle":
		port := util.FreePort()
		user := fmt.Sprintf("u_%v_%v", strings.Replace(util.MacAddr(), ":", "", -1), port)
		fmt.Fprintf(os.Stdout, "local: %v user: %v\n", port, user)

		go tunnel.TunClient(append(args, fmt.Sprintf("localhost:%v:localhost:%v", port, huskiePort)))
		//
		for {
			rc := ssh.Client([]string{"--p", fmt.Sprintf("%v", port), "--i", os.Getenv("HUSKIE_IDENTITY"), user + "@localhost"})
			sleep(rc)
		}
	case "ssh":
		port := util.FreePort()
		fmt.Fprintf(os.Stdout, "local: %v remote: %v\n", port, remotePort)

		go tunnel.TunClient(append(args, fmt.Sprintf("localhost:%v:localhost:%v", port, remotePort)))
		//
		for {
			rc := ssh.Client([]string{"--p", fmt.Sprintf("%v", port), "--i", os.Getenv("HUSKIE_IDENTITY"), "ubuntu@localhost"})
			if rc == 0 {
				os.Exit(0)
			}
			sleep(rc)
		}
	case "sshd":
		port := util.FreePort()
		servicePort := util.FreePort()
		fmt.Fprintf(os.Stdout, "local: %v remote: %v service: %v\n", port, remotePort, servicePort)

		go ssh.Server(servicePort)
		go tunnel.TunClient(append(args, fmt.Sprintf("localhost:%v:localhost:%v", port, serverPort)))

		for {
			rc := rp.Client(fmt.Sprintf(rpc, port, servicePort, remotePort))
			sleep(rc)
		}
	default:
		fmt.Fprintf(os.Stderr, help)
		os.Exit(1)
	}
}

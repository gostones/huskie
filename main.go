package main

import (
	"flag"
	"fmt"
	"github.com/gostones/huskie/bot"
	"github.com/gostones/huskie/chat"
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
var sshrps = `
[common]
bind_port = %v
`

//server_port, local_port, remote_port
var sshrpc = `
[common]
server_addr = localhost
server_port = %v
http_proxy =

[ssh]
type = tcp
local_ip = localhost
local_port = %v
remote_port = %v
`

//
var webrps = `
[common]
bind_port = %v
`
var webrpc = `
[common]
server_addr = localhost
server_port = %v
http_proxy =

[web]
type = tcp
local_ip = localhost
local_port = %v
remote_port = 58080
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

	//remotePort := 6000

	//sshPort := 7022
	//webport := 7080

	sleep := util.BackoffDuration()

	//
	switch subcmd {
	case "harness":
		port := os.Getenv("HUSKIE_PORT")
		go chat.Server([]string{"--bind", ":" + port, "--identity", os.Getenv("HUSKIE_IDENTITY")})
		//go rp.Server(fmt.Sprintf(sshrps, sshPort))
		//go rp.Server(fmt.Sprintf(webrps, webport))
		//go ssh.Server(9022)

		tunnel.TunServer(os.Getenv("PORT"))
	case "whistle":
		lport := util.FreePort()
		user := genUser(lport)
		fmt.Fprintf(os.Stdout, "local: %v user: %v\n", lport, user)

		url := os.Getenv("HUSKIE_URL")
		rport := os.Getenv("HUSKIE_PORT")

		proxy := os.Getenv("http_proxy")
		remote := fmt.Sprintf("localhost:%v:localhost:%v", lport, rport)
		go tunnel.TunClient(proxy, url, remote)

		for {
			rc := ssh.Client([]string{"--p", fmt.Sprintf("%v", lport), "--i", os.Getenv("HUSKIE_IDENTITY"), user + "@localhost"})
			if rc == 0 {
				os.Exit(0)
			}
			sleep(rc)
		}
	case "pup":
		lport := util.FreePort()
		user := fmt.Sprintf("puppy%v", lport)
		fmt.Fprintf(os.Stdout, "local: %v user: %v\n", lport, user)

		url := os.Getenv("HUSKIE_URL")
		rport := os.Getenv("HUSKIE_PORT")

		proxy := os.Getenv("http_proxy")
		remote := fmt.Sprintf("localhost:%v:localhost:%v", lport, rport)
		go tunnel.TunClient(proxy, url, remote)

		for {
			rc := bot.Server(user, "localhost", lport)
			if rc == 0 {
				os.Exit(0)
			}
			sleep(rc)
		}
	//case "bash":
	//	rport := 9022
	//	port := util.FreePort()
	//	fmt.Fprintf(os.Stdout, "local: %v remote: %v\n", port, rport)
	//
	//	go tunnel.TunClient(append(args, fmt.Sprintf("localhost:%v:localhost:%v", port, rport)))
	//	//
	//	for {
	//		rc := ssh.Client([]string{"--p", fmt.Sprintf("%v", port), "--i", os.Getenv("HUSKIE_IDENTITY"), "localhost"})
	//		if rc == 0 {
	//			os.Exit(0)
	//		}
	//
	//		sleep(rc)
	//	}
	//case "mush":
	//	port := util.FreePort()
	//	fmt.Fprintf(os.Stdout, "local: %v remote: %v\n", port, remotePort)
	//
	//	go tunnel.TunClient(append(args, fmt.Sprintf("localhost:%v:localhost:%v", port, remotePort)))
	//	//
	//	for {
	//		rc := ssh.Client([]string{"--p", fmt.Sprintf("%v", port), "--i", os.Getenv("HUSKIE_IDENTITY"), "ubuntu@localhost"})
	//		if rc == 0 {
	//			os.Exit(0)
	//		}
	//
	//		sleep(rc)
	//	}
	//case "dog":
	//	port := util.FreePort()
	//	servicePort := util.FreePort()
	//	fmt.Fprintf(os.Stdout, "local: %v remote: %v service: %v\n", port, remotePort, servicePort)
	//
	//	go docker.Server([]string{"ssh2docker", "--bind", fmt.Sprintf(":%v", servicePort)})
	//	go tunnel.TunClient(append(args, fmt.Sprintf("localhost:%v:localhost:%v", port, serverPort)))
	//
	//	for {
	//		rc := rp.Client(fmt.Sprintf(sshrpc, port, servicePort, remotePort))
	//		sleep(rc)
	//	}
	//case "webc":
	//	port := util.FreePort()
	//	lport := 18080
	//	fmt.Fprintf(os.Stdout, "local: %v  local http: %v web service: %v\n", port, lport, webport)
	//
	//	//TODO go start rshiny/any web app
	//	go tunnel.TunClient(append(args, fmt.Sprintf("localhost:%v:localhost:%v", port, webport)))
	//
	//	for {
	//		rc := rp.Client(fmt.Sprintf(webrpc, port, lport))
	//		sleep(rc)
	//	}
	//case "ssh":
	//	port := util.FreePort()
	//	fmt.Fprintf(os.Stdout, "local: %v remote: %v\n", port, remotePort)
	//
	//	go tunnel.TunClient(append(args, fmt.Sprintf("localhost:%v:localhost:%v", port, remotePort)))
	//	//
	//	for {
	//		rc := ssh.Client([]string{"--p", fmt.Sprintf("%v", port), "--i", os.Getenv("HUSKIE_IDENTITY"), "localhost"})
	//		if rc == 0 {
	//			os.Exit(0)
	//		}
	//		sleep(rc)
	//	}
	//case "sshd":
	//	port := util.FreePort()
	//	servicePort := 22 //util.FreePort()
	//	//go ssh.Server(servicePort)
	//	fmt.Fprintf(os.Stdout, "local: %v remote: %v service: %v\n", port, remotePort, servicePort)
	//
	//	go tunnel.TunClient(append(args, fmt.Sprintf("localhost:%v:localhost:%v", port, sshPort)))
	//	for {
	//		rc := rp.Client(fmt.Sprintf(sshrpc, port, servicePort, remotePort))
	//		sleep(rc)
	//	}
	default:
		fmt.Fprintf(os.Stderr, help)
		os.Exit(1)
	}
}

func genUser(rand int) string {
	return fmt.Sprintf("u_%v_%v", strings.Replace(util.MacAddr(), ":", "", -1), rand)
}

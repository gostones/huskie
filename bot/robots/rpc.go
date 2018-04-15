package robots

import (
	"fmt"
	"github.com/gostones/huskie/rp"
	"github.com/gostones/huskie/util"
	"github.com/jpillora/chisel/client"
	"log"
	"os"
	"strconv"
	"time"
)

// RpcBot starts reverse proxy client
type RpcBot struct {
	url   string
	proxy string
}

func init() {
	url := os.Getenv("HUSKIE_URL")
	proxy := os.Getenv("http_proxy")

	RegisterRobot("rpc", func() (robot Robot) {
		return &RpcBot{
			url:   url,
			proxy: proxy,
		}
	})
}

// Run executes a command
func (b RpcBot) Run(c *Command) string {
	if len(c.Args) != 2 {
		return "missing ports: service remote"
	}

	sport, err := strconv.Atoi(c.Args[0])
	if err != nil {
		return fmt.Sprintf("%v", err)
	}

	rport, err := strconv.Atoi(c.Args[1])
	if err != nil {
		return fmt.Sprintf("%v", err)
	}

	go b.sshd(sport, rport)

	return fmt.Sprintf("Started service: %v remote: %v", sport, rport)
}

// Description describes what the robot does
func (b RpcBot) Description() string {
	return "<something>"
}

//server_port, local_port, remote_port
var rpc = `
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

func (b RpcBot) sshd(sport, rport int) {
	lport := util.FreePort()

	remote := fmt.Sprintf("localhost:%v:localhost:%v", lport, 8000)

	go b.tunClient(b.proxy, b.url, remote)

	fmt.Fprintf(os.Stdout, "service: %v remote: %v proxy: %v url: %v\n", sport, remote, b.proxy, b.url)

	sleep := util.BackoffDuration()

	for {
		rc := rp.Client(fmt.Sprintf(rpc, lport, sport, rport))
		if rc == 0 {
			return
		}
		sleep(rc)
	}
}

func (b RpcBot) tunClient(proxy string, url string, remote string) {

	keepalive := time.Duration(12 * time.Second)

	c, err := chclient.NewClient(&chclient.Config{
		Fingerprint: "",
		Auth:        "",
		KeepAlive:   keepalive,
		HTTPProxy:   proxy,
		Server:      url,
		Remotes:     []string{remote},
	})

	c.Debug = true

	defer c.Close()
	if err = c.Run(); err != nil {
		log.Println(err)
	}
}

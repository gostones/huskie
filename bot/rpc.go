package bot

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
	RegisterRobot("rpc", func() (robot Robot) {
		return &RpcBot{
			url:   HuskieUrl,
			proxy: ProxyUrl,
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

	go b.tun(sport, rport)

	return fmt.Sprintf("Started service: %v remote: %v", sport, rport)
}

// Description describes what the robot does
func (b RpcBot) Description() string {
	return "service_port remote_port"
}

func (b RpcBot) tun(sport, rport int) {
	lport := util.FreePort()

	remote := fmt.Sprintf("localhost:%v:localhost:%v", lport, 8000)

	go b.tunClient(b.proxy, b.url, remote)

	fmt.Fprintf(os.Stdout, "service: %v remote: %v proxy: %v url: %v\n", sport, remote, b.proxy, b.url)

	sleep := util.BackoffDuration()

	for {
		rc := rp.Client(fmt.Sprintf(rpc, lport, rport, sport, rport))
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

package main

import (
	"flag"
	"fmt"
	"github.com/gostones/huskie/bot"
	"github.com/gostones/huskie/chat"
	"github.com/gostones/huskie/rp"
	"github.com/gostones/huskie/ssh"
	"github.com/gostones/huskie/tunnel"
	"github.com/gostones/huskie/util"
	"os"
	"strconv"
)

//
var help = `
	Usage: huskie [command] [--help]

	Commands:
		harness - server mode
		pup     - worker
		mush    - control agent
`

//
func main() {

	flag.Bool("help", false, "")
	flag.Bool("h", false, "")
	flag.Usage = func() {}
	flag.Parse()

	args := flag.Args()

	subcmd := ""
	if len(args) > 0 {
		subcmd = args[0]
		args = args[1:]
	}

	//
	switch subcmd {
	case "harness":
		harness(args)
	case "pup":
		puppy(args)
	case "mush":
		mush(args)
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, help)
	os.Exit(1)
}

//func genUser(rand int) string {
//	return fmt.Sprintf("u_%v_%v", strings.Replace(util.MacAddr(), ":", "", -1), rand)
//}

var rps = `
[common]
bind_port = %v
`
var huskiePort = 2022

func harness(args []string) {
	flags := flag.NewFlagSet("server", flag.ContinueOnError)

	//tunnel port
	bind := flags.Int("bind", parseInt(os.Getenv("PORT"), 8080), "")

	//chat port
	port := flags.Int("port", parseInt(os.Getenv("HUSKIE_PORT"), huskiePort), "")
	ident := flags.String("identity", os.Getenv("HUSKIE_IDENTITY"), "")

	v := flags.Bool("verbose", false, "")

	rport := flags.Int("rps", 8000, "")
	sport := flags.Int("ssh", 8022, "")

	flags.Parse(args)

	//
	args = []string{}
	args = append(args, "--bind", fmt.Sprintf(":%v", *port))

	if *ident == "" {
		*ident = "host_key"
		util.RsaKeyPair(*ident)
	}
	args = append(args, "--identity", *ident)

	if *v {
		args = append(args, "-v")
	}

	//
	go chat.Server(args)

	go rp.Server(fmt.Sprintf(rps, *rport))
	go ssh.Server(*sport, "bash")

	tunnel.TunServer(fmt.Sprintf("%v", *bind))
}

func mush(args []string) {
	flags := flag.NewFlagSet("connect", flag.ContinueOnError)

	port := flags.Int("port", parseInt(os.Getenv("HUSKIE_PORT"), huskiePort), "")
	ident := flags.String("identity", os.Getenv("HUSKIE_IDENTITY"), "")
	url := flags.String("url", os.Getenv("HUSKIE_URL"), "")
	proxy := flags.String("proxy", "", "")

	flags.Parse(args)

	if *url == "" {
		usage()
	}

	if *proxy == "" {
		*proxy = os.Getenv("http_proxy")
	}

	if *ident == "" {
		*ident = "host_key"
		util.RsaKeyPair(*ident)
	}

	//
	lport := util.FreePort()

	fmt.Fprintf(os.Stdout, "local: %v remote: %v\n", lport, *port)

	//
	remote := fmt.Sprintf("localhost:%v:localhost:%v", lport, *port)
	go tunnel.TunClient(*proxy, *url, remote)

	args = []string{"--p", fmt.Sprintf("%v", lport), "--i", *ident, "localhost"}
	sleep := util.BackoffDuration()
	for {
		rc := ssh.Client(args)
		if rc == 0 {
			os.Exit(0)
		}
		sleep(rc)
	}
}

func puppy(args []string) {
	flags := flag.NewFlagSet("puppy", flag.ContinueOnError)

	port := flags.Int("port", parseInt(os.Getenv("HUSKIE_PORT"), 2022), "")
	url := flags.String("url", os.Getenv("HUSKIE_URL"), "")
	proxy := flags.String("proxy", os.Getenv("http_proxy"), "")
	config := flags.String("config", "", "")

	flags.Parse(args)

	if *url == "" {
		usage()
	}

	//
	lport := util.FreePort()
	user := fmt.Sprintf("puppy%v", lport)

	fmt.Fprintf(os.Stdout, "local: %v user: %v\n", lport, user)
	fmt.Fprintf(os.Stdout, "config: %v", config)

	remote := fmt.Sprintf("localhost:%v:localhost:%v", lport, *port)
	go tunnel.TunClient(*proxy, *url, remote)

	sleep := util.BackoffDuration()

	for {
		rc := bot.Server(*proxy, *url, user, "localhost", lport)
		sleep(rc)
	}
}

func parseInt(s string, v int) int {
	if s == "" {
		return v
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		i = v
	}
	return i
}

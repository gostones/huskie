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
		pup     - worker mode (docker instance)
		whistle - messaging service
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
	case "whistle":
		whistle(args)
	case "pup":
		puppy(args)
	case "mush":
		mush(args)
	default:
		fmt.Fprintf(os.Stderr, help)
		os.Exit(1)
	}
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

	b := flags.Int("b", -1, "")
	bind := flags.Int("bind", -1, "")

	p := flags.Int("p", -1, "")
	port := flags.Int("port", -1, "")

	i := flags.String("i", "", "")
	ident := flags.String("identity", "", "")

	v := flags.Bool("v", false, "")

	rport := flags.Int("rps", 8000, "")

	sport := flags.Int("ssh", 8022, "")

	flags.Parse(args)

	//
	args = []string{}

	if *bind == -1 {
		*bind = *b
	}
	if *bind == -1 {
		*bind = parseInt(os.Getenv("PORT"), 8080)
	}

	if *port == -1 {
		*port = *p
	}
	if *port == -1 {
		*port = parseInt(os.Getenv("HUSKIE_PORT"), huskiePort)
	}

	args = append(args, "--bind", fmt.Sprintf(":%v", *port))

	if *ident == "" {
		*ident = *i
	}
	if *ident == "" {
		*ident = os.Getenv("HUSKIE_IDENTITY")
	}
	if *ident == "" {
		*ident = "host_key"
		util.RsaKeyPair(*ident)
	}

	args = append(args, "--identity", *ident)

	if *v {
		args = append(args, "-v")
	}

	go chat.Server(args)

	go rp.Server(fmt.Sprintf(rps, *rport))
	go ssh.Server(*sport, "bash")

	tunnel.TunServer(fmt.Sprintf("%v", *bind))
}

func whistle(args []string) {
	connect(args)
}

func mush(args []string) {
	connect(args)
}

func connect(args []string) {
	flags := flag.NewFlagSet("connect", flag.ContinueOnError)

	p := flags.Int("p", -1, "")
	port := flags.Int("port", -1, "")

	i := flags.String("i", "", "")
	ident := flags.String("identity", "", "")

	u := flags.String("u", "", "")
	url := flags.String("url", "", "")

	proxy := flags.String("proxy", "", "")

	flags.Parse(args)

	if *url == "" {
		*url = *u
	}
	if *url == "" {
		*url = os.Getenv("HUSKIE_URL")
	}
	if *url == "" {
		*url = "http://localhost:8080/tunnel"
	}

	if *proxy == "" {
		*proxy = os.Getenv("http_proxy")
	}

	if *port == -1 {
		*port = *p
	}
	if *port == -1 {
		*port = parseInt(os.Getenv("HUSKIE_PORT"), huskiePort)
	}

	if *ident == "" {
		*ident = *i
	}
	if *ident == "" {
		*ident = os.Getenv("HUSKIE_IDENTITY")
	}
	if *ident == "" {
		*ident = "host_key"
		util.RsaKeyPair(*ident)
	}

	//
	lport := util.FreePort()

	fmt.Fprintf(os.Stdout, "local: %v remote: %v\n", lport, *port)

	remote := fmt.Sprintf("localhost:%v:localhost:%v", lport, *port)
	go tunnel.TunClient(*proxy, *url, remote)

	//
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

	p := flags.Int("p", -1, "")
	port := flags.Int("port", -1, "")

	u := flags.String("u", "", "")
	url := flags.String("url", "", "")

	proxy := flags.String("proxy", "", "")

	flags.Parse(args)

	if *url == "" {
		*url = *u
	}
	if *url == "" {
		*url = os.Getenv("HUSKIE_URL")
	}
	if *url == "" {
		*url = "http://localhost:8080/tunnel"
	}

	if *proxy == "" {
		*proxy = os.Getenv("http_proxy")
	}

	//
	if *port == -1 {
		*port = *p
	}
	if *port == -1 {
		*port = parseInt(os.Getenv("HUSKIE_PORT"), 2022)
	}

	lport := util.FreePort()
	user := fmt.Sprintf("puppy%v", lport)
	fmt.Fprintf(os.Stdout, "local: %v user: %v\n", lport, user)

	remote := fmt.Sprintf("localhost:%v:localhost:%v", lport, *port)
	go tunnel.TunClient(*proxy, *url, remote)

	sleep := util.BackoffDuration()

	for {
		rc := bot.Server(user, "localhost", lport)
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

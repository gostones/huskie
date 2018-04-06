package docker

import (
	"log/syslog"
	"net"
	"os"
	"strings"

	"github.com/apex/log"
	"github.com/apex/log/handlers/multi"
	"github.com/apex/log/handlers/text"
	"github.com/codegangsta/cli"
	"github.com/moul/ssh2docker"
	"github.com/moul/ssh2docker/pkg/sysloghandler"
)

var VERSION string

// Default key to ease getting started with the server
// You can easily use your own key by setting up
// `--host-key=/path/to/id_rsa`.
// See `man 1 ssh-keygen`.
const DefaultHostKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA0BGfZSFn5ueRzMGPnd4+QkbrJ5vRmRXdg0D3ukSxFQC+QCXM
5lVYDtqp6DSsiIIr3PB0n94onebdDK7763RO1/fJP0ZBN3Ih1q0oQ9llrq7kuOd/
ttjNviu9KAkVQLcfR6zxttIPu/xnwkn7Y0pNOpn6ytjA2whemEKTAyskLSNVBqtW
r2TY/am7aXYG+1HSkbfSTSKI4ekzHAFLAZGK1q4FDOMAs6kC4IEmop1T3O2LvPBF
QzTt2WT0kph4+4saMqo4yoKEcKbdnWRkZul1YOcVyJReFX78fCKGo9tVjwtHHa3C
98WgjUiAXN2boGY2tPqk5vrTVmB69CJ6reezKQIDAQABAoIBAEBtRILnDio0iDPz
t4m1mGejWAtCt2sElzueMVcPEBolycNJMSIdSRAIa1YIgWgfjn9yQVqDSuZh5w6X
XFAzCnrbMgiSs3z8rTexFGe1+ENXymDq5ePzS/nXx1GPRnJsgZYLGil28AJQjLxf
diTvi+xaY4rOBSGNfOT+sFDp2eDTofTbxidgDzEJe0tWMi9QHs5NkyERYO7cpd6I
uXe0QLMP0aBzMKH4BFMiyBcY2gxYxqr1rC7YmEmn6M2HKxsHjGJ9K7msyJ0CdXNk
tiqwi3++T3jOcOz3u1t921/wUb4P7TJmerduEt5fm6wJVPCwZdQBcmoMznC0FGb2
5UwxeaUCgYEA7PTCQi7nxJB56CTxgM6m3D/+Lzdq1SBgUt3kzKPtJtDAfSCBd4hL
NYNyJ2WqJBdoclY+FQq85Rn2EdZY7Baol3WK26y+xWgXrICzPAtBvNByrr7eFpJS
2c35PrEuvJvsx0yn2HM4a8JYAm2yq+iJuW0aUNquhFLECvivsCfmXjcCgYEA4MqG
Rxy2Xu47UELfG3GONe6qd9gv2WWMzXtGXQfzYcYQMFUtfAj04/5ER92BEVqVTARB
ZOhdgLPTHsP4FwcjsCpjZQ2cW6QXUQfGJGrCWWd+3tFSv7ekrp4mXl4iXT/zJnWB
BrbABDS8qwz2+YhV6ql2wlYbVI7wdAf4tev2yZ8CgYBQNWmsTYRWnTEmy5qUJ1+E
HoVEJlYbXqI8arAQNU0JXpBJyr8IXzJWIvB5NYiqPuI0Ec1iAgh+5JLO5ueiwui+
nCMsyQSqfdnFoqsJICZYa5bmX+V9bnptD7PW7NMNNRqpO+F0+0uV7mssJ0XbuxMj
mTLXO67nS7zgmd2em2L3cQKBgAFYNMVoHo8izagFPmBjpX4dF1fwKxkZymXQPvN/
gK0tChu/5q2/P/e9JZtob8UyzYHO5LU9zpFegfzFH07D9CqxljachjrmGF2btkux
d8ghHlkm11/eMVX6DDC0T3BPWZz5RvRLU4qy5g3/3dpQPnNQ4Cz5ZuBymm2XPp2X
87nxAoGBAIh+WKnvopNektUbJX7hDE1HBVVaFfM3VffNDfBYF9ziPotRqY+th30N
Nn/aVy1yLb3M0eZCFD9A9/NE23kXAIFMAjM44NpqtXU/Z6Dmc+qB0Lk+KSL72cQL
hLP2sDRnSOAAGEYHcN75mvdrKahDRQXD8l0FIRUn/BNTlob8g5Tp
-----END RSA PRIVATE KEY-----`

func Server(args []string) {
	app := cli.NewApp()
	app.Name = "ssh2docker" //path.Base(os.Args[0])
	//app.Author = "Manfred Touron"
	//app.Email = "https://github.com/moul/ssh2docker"
	app.Version = VERSION
	app.Usage = "SSH portal to Docker containers"

	app.Before = hookBefore

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, V",
			Usage: "Enable verbose mode",
		},
		cli.StringFlag{
			Name:  "syslog-server",
			Usage: "Configure a syslog server, i.e: udp://localhost:514",
		},
		cli.StringFlag{
			Name:  "bind, b",
			Value: ":2222",
			Usage: "Listen to address",
		},
		cli.StringFlag{
			Name:  "host-key, k",
			Usage: "Path or complete SSH host key to use, use 'system' for keys in /etc/ssh",
			Value: "built-in",
		},
		cli.StringFlag{
			Name:  "allowed-images",
			Usage: "List of allowed images, i.e: alpine,ubuntu:trusty,1cf3e6c",
			Value: "",
		},
		cli.StringFlag{
			Name:  "shell",
			Usage: "Default shell",
			Value: "/bin/sh",
		},
		cli.StringFlag{
			Name:  "docker-run-args",
			Usage: "'docker run' arguments",
			Value: "-i {{if .UseTTY}} -t {{end}} --rm",
		},
		cli.StringFlag{
			Name:  "docker-exec-args",
			Usage: "'docker exec' arguments",
			Value: "-i {{if .UseTTY}} -t {{end}}",
		},
		cli.BoolFlag{
			Name:  "no-join",
			Usage: "Do not join existing containers, always create new ones",
		},
		cli.BoolFlag{
			Name:  "clean-on-startup",
			Usage: "Cleanup Docker containers created by ssh2docker on start",
		},
		cli.StringFlag{
			Name:  "password-auth-script",
			Usage: "Password auth hook file",
		},
		cli.StringFlag{
			Name:  "publickey-auth-script",
			Usage: "Public-key auth hook file",
		},
		cli.StringFlag{
			Name:  "local-user",
			Usage: "If setted, you can spawn a local shell (not withing docker) by SSHing to this user",
		},
		cli.StringFlag{
			Name:  "banner",
			Usage: "Display a banner on connection",
		},
	}

	app.Action = Action

	app.Run(args)
}

func hookBefore(c *cli.Context) error {
	level := log.InfoLevel
	syslogLevel := syslog.LOG_INFO
	if c.Bool("verbose") {
		level = log.DebugLevel
		syslogLevel = syslog.LOG_DEBUG
	}
	log.SetLevel(level)
	log.SetHandler(text.New(os.Stderr))

	if c.String("syslog-server") != "" {
		server := strings.Split(c.String("syslog-server"), "://")

		if server[0] == "unix" {
			log.SetHandler(multi.New(text.New(os.Stderr), sysloghandler.New("", "", syslogLevel, "")))
		} else {
			if len(server) != 2 {
				log.Fatal("invalid syslog parameter")
			}
			log.SetHandler(multi.New(text.New(os.Stderr), sysloghandler.New(server[0], server[1], syslogLevel, "")))
		}
	}
	return nil
}

// Action is the default cli action to execute
func Action(c *cli.Context) {
	// Initialize the SSH server
	server, err := ssh2docker.NewServer()
	if err != nil {
		log.Fatalf("Cannot create server: %v", err)
	}

	// Restrict list of allowed images
	if c.String("allowed-images") != "" {
		server.AllowedImages = strings.Split(c.String("allowed-images"), ",")
	}

	// Configure server
	server.DefaultShell = c.String("shell")
	server.DockerRunArgsInline = c.String("docker-run-args")
	server.DockerExecArgsInline = c.String("docker-exec-args")
	server.NoJoin = c.Bool("no-join")
	server.CleanOnStartup = c.Bool("clean-on-startup")
	server.PasswordAuthScript = c.String("password-auth-script")
	server.PublicKeyAuthScript = c.String("publickey-auth-script")
	server.LocalUser = c.String("local-user")
	server.Banner = c.String("banner")

	// Register the SSH host key
	hostKey := c.String("host-key")
	switch hostKey {
	case "built-in":
		hostKey = DefaultHostKey
	case "system":
		hostKey = "/etc/ssh/ssh_host_rsa_key"
	}
	err = server.AddHostKey(hostKey)
	if err != nil {
		log.Fatalf("Cannot add host key: %v", err)
	}

	// Bind TCP socket
	bindAddress := c.String("bind")
	listener, err := net.Listen("tcp", bindAddress)
	if err != nil {
		log.Fatalf("Failed to start listener on %q: %v", bindAddress, err)
	}
	log.Infof("Listening on %q", bindAddress)

	// Initialize server
	if err = server.Init(); err != nil {
		log.Fatalf("Failed to initialize the server: %v", err)
	}

	// Accept new clients
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorf("Accept failed: %v", err)
			continue
		}
		go server.Handle(conn)
	}
}

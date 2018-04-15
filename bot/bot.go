package bot

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gostones/huskie/bot/robots"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"strings"
	"time"
)

type ChatMessage struct {
	Type    string                 `json:"type"`
	To      string                 `json:"to"`
	From    string                 `json:"from"`
	Msg     string                 `json:"msg"`
	Content map[string]interface{} `json:"content"`
}

var active = false

// Bot runs the bot
func Bot(user string, addr string) error {
	conn, err := dial(addr, user)
	if err != nil {
		return err
	}

	session, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	in, err := session.StdinPipe()
	if err != nil {
		return err
	}

	out, err := session.StdoutPipe()
	if err != nil {
		return err
	}

	if err := session.Shell(); err != nil {
		return err
	}

	err = session.RequestPty("xterm", 80, 40, ssh.TerminalModes{})
	if err != nil {
		return err
	}

	delay := 5*time.Second

	go func() {
		time.Sleep(delay)
		in.Write([]byte("/me now active\r\n"))
		active = true
	}()

	//in.Write([]byte("/theme mono\r\n"))

	//go func() {
	//	for {
	//		in.Write([]byte("/motd\r\n"))
	//		time.Sleep(*check)
	//	}
	//}()

	scanner := bufio.NewScanner(out)

	for scanner.Scan() {
		line := scanner.Text()
		if err != nil {
			return err
		}
		fmt.Println(line)

		cm := ChatMessage{}
		err := json.Unmarshal([]byte(line), &cm)
		fmt.Printf("invalid json: %v", err)
		if err == nil {
			cmdline := strings.Split(cm.Msg, " ")
			cmd := &robots.Command{
				From:    cm.From,
				Command: cmdline[0],
				Args:    cmdline[1:],
			}

			robot, err := getRobot(cmd.Command)
			if err != nil {
				continue
			}

			if !active {
				continue
			}

			if response := robot.Run(cmd); response != "" {
				reply(in, fmt.Sprintf(`{ "to": "%s", "msg": "%s" }`, cm.From, response))
			}
		}

		//if strings.Contains(line, " "+*user+": ") {
		//	cmd, err := parseLine(line)
		//	if err == nil {
		//		robot, err := getRobot(cmd.Command)
		//		if err != nil {
		//			continue
		//		}
		//
		//		if !active {
		//			continue
		//		}
		//
		//		if response := robot.Run(cmd); response != "" {
		//			reply(in, fmt.Sprintf("%s %s", cmd.From, response))
		//		}
		//	}
		//}
	}

	return errors.New("ERROR")
}

//
//func parseLine(line string) (*robots.Command, error) {
//	fields := strings.Fields(line)
//
//	if len(fields) < 4 {
//		return nil, errors.New("not enough fields in line")
//	}
//
//	fromFields := strings.Split(fields[1], controlCodeString)
//	if len(fromFields) < 2 {
//		return nil, errors.New("not enough fields in line")
//	}
//	from := fromFields[1]
//
//	if len(fields) < 4 {
//		return nil, errors.New("not enough fields in line")
//	}
//
//	command := strings.TrimRight(fields[3], "\a")
//
//	args := []string{}
//
//	if len(fields) > 4 {
//		for _, f := range fields[4:] {
//			args = append(args, strings.TrimRight(f, "\a"))
//		}
//	}
//
//	if active {
//		fmt.Printf("%#v\n", args)
//	}
//
//	cmd := robots.Command{
//		From:    from,
//		Command: command,
//		Args:    args,
//	}
//
//	return &cmd, nil
//}

func reply(in io.WriteCloser, s string) {
	in.Write([]byte(s + "\r\n"))
}

func dial(addr, user string) (*ssh.Client, error) {
	key, err := MakeKey()
	if err != nil {
		return nil, err
	}

	return ssh.Dial("tcp", addr, &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	})
}

func getRobot(command string) (robots.Robot, error) {
	if robotInitFunction, ok := robots.Robots[command]; ok {
		return robotInitFunction(), nil
	}

	return nil, fmt.Errorf("unknown robot: %s", command)
}
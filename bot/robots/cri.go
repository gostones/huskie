package robots

import (
	"fmt"
	"github.com/gostones/huskie/docker"
	"strconv"
)

// Bot for container runtime interface
type CriBot struct{}

func init() {
	RegisterRobot("cri", func() (robot Robot) {
		return new(CriBot)
	})
}

// Run executes a command
func (b CriBot) Run(c *Command) string {
	if len(c.Args) == 0 {
		return "missing port"
	}

	port, err := strconv.Atoi(c.Args[0])
	if err != nil {
		return fmt.Sprintf("%v", err)
	}

	go docker.Server([]string{"ssh2docker", "--bind", fmt.Sprintf(":%v", port)})

	return fmt.Sprintf("cri started at %v", port)
}

// Description describes what the robot does
func (b CriBot) Description() string {
	return "<something>"
}

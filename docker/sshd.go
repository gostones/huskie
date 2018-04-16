package docker

import (
	"context"
	"fmt"
	"io"
	"log"

	"flag"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/gliderlabs/ssh"
	"github.com/gostones/huskie/util"
	"strings"
)

func Server(port int) {
	ssh.Handle(func(sess ssh.Session) {
		hport := util.FreePort()
		args := sess.Command()

		log.Printf("port: %v user: %v cmd: %v env: %v\n", hport, sess.User(), args, sess.Environ())

		//args
		flags := flag.NewFlagSet("docker", flag.ContinueOnError)

		port := flags.String("p", "8080", "")
		vol := flags.String("v", "", "")
		env := flags.String("e", "", "")

		image := args[0]
		flags.Parse(args[1:])

		vols := []string{}
		if *vol != "" {
			vols = []string{*vol}
		}

		envs := []string{}
		if *env != "" {
			envs = strings.Split(*env, ",")
		}

		cmd := flags.Args()

		log.Printf("image: %v vol: %v env: %v cmd: %v\n", image, vols, envs, cmd)

		//
		hcfg := &container.HostConfig{
			Binds: vols,
			PortBindings: nat.PortMap{
				nat.Port(fmt.Sprintf("%v/tcp", *port)): []nat.PortBinding{
					{
						HostIP:   "0.0.0.0",
						HostPort: fmt.Sprintf("%v", hport),
					},
				},
			},
		}

		//
		_, _, isTty := sess.Pty()
		cfg := &container.Config{
			User:  sess.User(),
			Image: image,
			//Cmd:          cmd,
			Env:          envs,
			Tty:          isTty,
			OpenStdin:    true,
			AttachStderr: true,
			AttachStdin:  true,
			AttachStdout: true,
			StdinOnce:    true,
			Volumes:      make(map[string]struct{}),
		}
		status, cleanup, err := dockerRun(cfg, hcfg, sess)
		defer cleanup()
		if err != nil {
			fmt.Fprintln(sess, err)
			log.Println(err)
		}
		sess.Exit(int(status))
	})

	log.Printf("starting ssh server on port %v...\n", port)
	log.Fatal(ssh.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

func dockerRun(cfg *container.Config, hcfg *container.HostConfig, sess ssh.Session) (status int64, cleanup func(), err error) {
	docker, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	status = 255
	cleanup = func() {}
	ctx := context.Background()
	res, err := docker.ContainerCreate(ctx, cfg, hcfg, nil, "")
	if err != nil {
		return
	}
	cleanup = func() {
		docker.ContainerRemove(ctx, res.ID, types.ContainerRemoveOptions{})
	}
	opts := types.ContainerAttachOptions{
		Stdin:  cfg.AttachStdin,
		Stdout: cfg.AttachStdout,
		Stderr: cfg.AttachStderr,
		Stream: true,
	}
	stream, err := docker.ContainerAttach(ctx, res.ID, opts)
	if err != nil {
		return
	}
	cleanup = func() {
		docker.ContainerRemove(ctx, res.ID, types.ContainerRemoveOptions{})
		stream.Close()
	}

	outputErr := make(chan error)

	go func() {
		var err error
		if cfg.Tty {
			_, err = io.Copy(sess, stream.Reader)
		} else {
			_, err = stdcopy.StdCopy(sess, sess.Stderr(), stream.Reader)
		}
		outputErr <- err
	}()

	go func() {
		defer stream.CloseWrite()
		io.Copy(stream.Conn, sess)
	}()

	err = docker.ContainerStart(ctx, res.ID, types.ContainerStartOptions{})
	if err != nil {
		return
	}
	if cfg.Tty {
		_, winCh, _ := sess.Pty()
		go func() {
			for win := range winCh {
				err := docker.ContainerResize(ctx, res.ID, types.ResizeOptions{
					Height: uint(win.Height),
					Width:  uint(win.Width),
				})
				if err != nil {
					log.Println(err)
					break
				}
			}
		}()
	}
	resultC, errC := docker.ContainerWait(ctx, res.ID, container.WaitConditionNotRunning)
	select {
	case err = <-errC:
		return
	case result := <-resultC:
		status = result.StatusCode
	}
	err = <-outputErr
	return
}

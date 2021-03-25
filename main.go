package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gkawamoto/rediscraft/config"
	"github.com/gkawamoto/rediscraft/server"
	"github.com/tidwall/redcon"
)

func main() {
	conf, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	ctx := contextWithSignal(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := &mux{&sync.Mutex{}, make(chan []byte), nil}

	errchan := make(chan error)

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		err := serveRedis(ctx, conf, mux)
		if err != nil {
			errchan <- err
		}
		wg.Done()
	}()
	go func() {
		defer cancel()
		err := serveMinecraft(ctx, conf, mux)
		if err != nil {
			errchan <- err
		}
		wg.Done()
	}()

	go func() {
		wg.Wait()
		close(errchan)
	}()

	err = <-errchan
	if err != nil {
		log.Fatal(err)
	}
}

func contextWithSignal(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
		<-c
		cancel()
	}()
	return ctx
}

type mux struct {
	readerMutex *sync.Mutex

	output   chan []byte
	commands []string
}

func (m *mux) AddCommand(command string) {
	m.readerMutex.Lock()
	m.commands = append(m.commands, command)
	m.readerMutex.Unlock()
}

func (m *mux) Read(p []byte) (int, error) {
	var commands []string
	m.readerMutex.Lock()
	if len(m.commands) > 0 {
		commands = m.commands
		m.commands = nil
	}
	m.readerMutex.Unlock()

	if len(commands) == 0 {
		return 0, nil
	}

	data := strings.Join(commands, "\r\n") + "\r\n"
	n := copy(p, data)
	return n, nil
}

func (m *mux) Write(p []byte) (int, error) {
	var data = make([]byte, len(p))
	copy(data, p)
	go func() {
		select {
		case m.output <- data:
		default:
		}
	}()

	return len(p), nil
}

func serveRedis(ctx context.Context, conf config.Config, m *mux) error {
	serverHandler := server.New(m, conf)
	server := redcon.NewServer(conf.RedisAddr, serverHandler.Handle, serverHandler.Connect, serverHandler.Disconnect)
	errchan := make(chan error)
	go func() {
		errchan <- server.ListenAndServe()
	}()

	for {
		select {
		case <-ctx.Done():
			server.Close()
			return nil
		case data := <-m.output:
			serverHandler.Output("stdout", data)
		case err := <-errchan:
			return err
		}
	}
}

func serveMinecraft(ctx context.Context, conf config.Config, m *mux) error {
	cmd := exec.Command(
		"java",
		fmt.Sprintf("-Xms%s", conf.MinecraftMemory),
		fmt.Sprintf("-Xmx%s", conf.MinecraftMemory),
		"-jar", conf.MinecraftJAR,
		"-nogui",
	)
	cmd.Stdout = io.MultiWriter(os.Stdout, m)
	cmd.Stderr = io.MultiWriter(os.Stderr, m)
	cmd.Stdin = m
	cmd.Dir = conf.MinecraftFolder
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	errchan := make(chan error)
	go func() {
		err := cmd.Start()
		if err != nil {
			errchan <- err
			return
		}

		_, err = cmd.Process.Wait()
		errchan <- err
	}()

	select {
	case <-ctx.Done():
		exitCode := cmd.ProcessState.ExitCode()
		log.Printf("trying to stop the server gracefully (current exit code: %d)", exitCode)
		if exitCode == -1 {
			// Issuing a /stop command to try and end the server gracefully
			m.AddCommand("/stop")
			select {
			case <-time.After(time.Second * 30):
				_ = cmd.Process.Kill()
				return fmt.Errorf("timed out exiting minecraft server after 30s, killed the process")
			case err := <-errchan:
				return err
			}
		}
		return nil
	case err := <-errchan:
		return err
	}
}

package main

import (
	"bufio"
	"log"
	"net/rpc/jsonrpc"
	"os/exec"
	"syscall"
	"time"

	"github.com/natefinch/pie"
	"github.com/pebbe/zmq4"
	"gopkg.in/mgo.v2/bson"
)

const defaultFailedCode = 1

var returncode string
var cmd *exec.Cmd

type Message struct {
	Plugin    string            `json:"plugin"`
	Timeout   time.Duration     `json:"timeout"`
	Args      map[string]string `json:"args"`
	Command   string            `json:"command"`
	StdSocket string            `json:"stdsocket"`
}

func main() {
	log.SetPrefix("[command plugin log] ")

	p := pie.NewProvider()
	if err := p.RegisterName("Plugin", api{}); err != nil {
		log.Fatalf("failed to register Plugin: %s", err)
	}
	p.ServeCodec(jsonrpc.NewServerCodec)
}

type api struct{}

func (api) Init(data *Message, response *bson.M) error {
	if data == nil {
		return nil
	}

	stdoutchan, err := zmq4.NewSocket(zmq4.PAIR)
	if err != nil {
		return err
	}
	err = stdoutchan.Connect(data.StdSocket)
	if err != nil {
		return err
	}
	defer stdoutchan.Close()

	command := data.Command

	if len(command) == 0 {
		stdoutchan.SendMessage("No command specified. Aborting.")
		stdoutchan.SendMessage("DONE", 500)
		return nil
	}

	cmd = exec.Command("bash", "-c", command)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Error setting up stdoutpipe: %s", err)
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("Error setting up stderrpipe: %s", err)
		return err
	}

	// start the command after having set up the pipe
	if err := cmd.Start(); err != nil {
		log.Printf("Error executing the command: %s", err)
		return err
	}
	log.Printf("Started process with pid %d\n", cmd.Process.Pid)

	var output []string
	// read command's stdout line by line
	bufout := bufio.NewScanner(stdout)
	buferr := bufio.NewScanner(stderr)

	go func() error {
		for buferr.Scan() {
			line := buferr.Text()
			output = append(output, line)
			stdoutchan.SendMessage(line)
		}
		return nil
	}()

	for bufout.Scan() {
		line := bufout.Text()
		output = append(output, line)
		stdoutchan.SendMessage(line)
	}

	var exitCode int
	if err := cmd.Wait(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		} else {
			exitCode = defaultFailedCode
		}
	} else {
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}

	*response = bson.M{
		"exitcode": exitCode,
		"result":   output,
	}

	// Provide your exitcode and optionally your result if you have any results to give back.
	// this *response data will be used in the future release of the worker.
	// however, you have to give the worker back a valid *response, even if it doesn't make any use of it, yet.
	*response = bson.M{
		"exitcode": 200,
		"result":   "",
	}

	// IMPORTANT:
	// we've reached the end of our plugin and everything went well. let's return with a 200 returncode.
	// again: you MUST close the stdoutchan channel with a "DONE" message, otherwise bad things will happen!
	stdoutchan.SendMessage("DONE", 200)

	return nil
}

func (api) Kill(data []string, response *[]string) error {
	err := cmd.Process.Kill()
	if err != nil {
		return err
	}
	log.Printf("Killed process with pid %d\n", cmd.Process.Pid)
	return nil
}

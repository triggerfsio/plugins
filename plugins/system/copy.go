package main

import (
	"log"
	"net/rpc/jsonrpc"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/natefinch/pie"
	"github.com/pebbe/zmq4"
	"gopkg.in/mgo.v2/bson"
)

const defaultFailedCode = 1

var returncode string
var stdio chan []string
var cmd *exec.Cmd

type Message struct {
	Plugin    string            `json:"plugin"`
	Timeout   time.Duration     `json:"timeout"`
	Args      map[string]string `json:"args"`
	Command   string            `json:"command"`
	StdSocket string            `json:"stdsocket"`
}

func main() {
	stdio = make(chan []string)
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

	command := data.Command
	args := data.Args

	stdoutchan, err := zmq4.NewSocket(zmq4.PAIR)
	if err != nil {
		return err
	}
	err = stdoutchan.Connect(data.StdSocket)
	if err != nil {
		return err
	}
	defer stdoutchan.Close()

	path := args["path"]

	if strings.HasSuffix(path, "/") {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			err := os.MkdirAll(path, 0755)
			if err != nil {
				stdoutchan.SendMessage("Couldn't create directory.")
				stdoutchan.SendMessage("CLOSE")
				return err
			}
		}
	} else {
		if _, err := os.Stat(filepath.Dir(path)); os.IsNotExist(err) {
			err := os.MkdirAll(filepath.Dir(path), 0755)
			if err != nil {
				stdoutchan.SendMessage("Couldn't create directory.")
				stdoutchan.SendMessage("CLOSE")
				return err
			}
		}
		file, err := os.Create(path)
		if err != nil {
			stdoutchan.SendMessage("Couldn't create file.")
			stdoutchan.SendMessage("CLOSE")
			return err
		}
		_, err = file.Write([]byte(command))
		if err != nil {
			stdoutchan.SendMessage("Couldn't write to file.")
			stdoutchan.SendMessage("CLOSE")
			return err
		}
	}

	*response = bson.M{
		"exitcode": 200,
		"result":   "OK",
	}

	stdoutchan.SendMessage("Successfully copied file.")

	// IMPORTANT:
	// you MUST close the stdoutchan channel with a "CLOSE" message, otherwise bad things will happen!
	stdoutchan.SendMessage("CLOSE")

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

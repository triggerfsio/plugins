package main

import (
	"log"
	"net/rpc/jsonrpc"
	"time"

	"github.com/natefinch/pie"
	"github.com/pebbe/zmq4"
	"gopkg.in/mgo.v2/bson"
)

type Message struct {
	Plugin    string            `json:"plugin"`
	Timeout   time.Duration     `json:"timeout"`
	Args      map[string]string `json:"args"`
	Command   string            `json:"command"`
	StdSocket string            `json:"stdsocket"`
}

func main() {
	// set prefix of module logger.
	// the logs in the terminal where the worker is running will be prefixed with this string.
	log.SetPrefix("[awesomeplugin log] ")

	// mandatory:
	// we initialize a new provider provided by the pie package
	p := pie.NewProvider()

	// we HAVE to register the name of this plugin with the string "Plugin"
	if err := p.RegisterName("Plugin", api{}); err != nil {
		log.Fatalf("failed to register Plugin: %s", err)
	}
	p.ServeCodec(jsonrpc.NewServerCodec)
}

type api struct{}

// NOTE: you HAVE to have a function called "Init". this is the entrance of the worker process calling your plugin
func (api) Init(data *Message, response *bson.M) error {
	// check if we receive data at all or return.
	// you can skip this if your plugin doesn't expect any input data
	if data == nil {
		return nil
	}

	command := data.Command
	_ = data.Args

	// init zmq std socket for realtime std IO communication back to the client.
	stdoutchan, err := zmq4.NewSocket(zmq4.PAIR)
	if err != nil {
		return err
	}
	err = stdoutchan.Connect(data.StdSocket)
	if err != nil {
		return err
	}
	defer stdoutchan.Close()

	// now do something with the command given to your by the client.
	log.Print(command)

	// occasionally update the client with information in realtime
	stdoutchan.SendMessage("look what I'm doing ma...")

	// Provide your exitcode and optionally your result if you have any results to give back.
	*response = bson.M{
		"exitcode": 200,
		"result":   "",
	}

	// IMPORTANT:
	// you MUST close the stdoutchan channel with a "CLOSE" message, otherwise bad things will happen!
	stdoutchan.SendMessage("CLOSE")

	return nil
}

// NOTE: you HAVE to have a function called "Kill"
// THIS FUNCTION HAS TO EXIST (you can change the function body. for example if you have no process to kill, just return)
func (api) Kill(data []string, response *[]string) error {
	// we have no kill method here
	return nil
}

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
	if data == nil {
		return nil
	}

	// MANDATORY
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

	// this is the payload the client has sent you
	command := data.Command

	// EXAMPLE: abort plugin if no command was specified. (you can skip this if your plugin doesn't expect any input data)
	// notice how we close the stdoutchan with "DONE" here. you MUST close the channel with a "DONE" message before you return!
	if len(command) == 0 {
		stdoutchan.SendMessage("No command specified. Aborting.")
		stdoutchan.SendMessage("DONE", 500)
		return nil
	}

	// EXAMPLE: CHECK FOR PLUGIN ARGUMENTS
	// these are the plugin arguments the client has sent to you
	// there can be multiple plugin args. only you (as the plugin developer) know what type of args you expect.
	// so check for existing args or notify client that you expect something.
	args := data.Args
	if _, ok := args["key1"]; !ok {
		// key1 is not there. tell the client and exit
		stdoutchan.SendMessage("Expecting plugin argument called key1. Aborting.")

		// we notify the worker that we are already done and want to return immediately.
		// again: you HAVE to send a "DONE" followed by a returncode here to return properly.
		stdoutchan.SendMessage("DONE", 500)
		return nil
	}

	// EXAMPLE: now do something with the command given to your by the client.
	log.Print(command)
	// occasionally update the client with information in realtime
	stdoutchan.SendMessage("look what I'm doing ma...")
	// more work
	stdoutchan.SendMessage("here is your realtime output")
	// more work
	stdoutchan.SendMessage("and another one")

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

// NOTE: you HAVE to have a function called "Kill"
// THIS FUNCTION HAS TO EXIST (you can change the function body. for example if you have no process to kill, just return)
// this Kill function will be called by the worker if a timeout has been reached.
// you can put your code here for proper cleanup of your plugin, to remove dirty files, or whatever you need to do to reach a clean state.
func (api) Kill(data []string, response *[]string) error {
	// we have nothing to do here. so just return.
	return nil
}

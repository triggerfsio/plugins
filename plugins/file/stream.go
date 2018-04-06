package main

import (
	"io"
	"log"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/triggerfsio/plugins"
)

type Command struct {
	Plugin *plugins.PluginWrapper
}

func main() {
	plugin := plugins.NewPlugin()
	plugin.Start(&Command{Plugin: plugin})
}

func (c *Command) Init(message *plugins.Message, resp *plugins.Response) error {
	err := c.Plugin.Open(message.Socket)
	if err != nil {
		return err
	}
	defer c.Plugin.Close()

	// let's use a function outside of *Command for demonstation purposes.
	err = run(c, message, resp)
	if err != nil {
		return err
	}
	return nil
}

func (c *Command) Kill(message *plugins.Message, resp *plugins.Response) error {
	return nil
}

const bufferSize = 1024

// this function has not the Command receiver by intention.
// just to show an example of how one can use an own function.
func run(c *Command, message *plugins.Message, resp *plugins.Response) error {
	if _, ok := message.Args["path"]; !ok {
		c.Plugin.Send("Please specify a path in your plugin arguments.")
		return nil
	}

	file, err := os.Open(message.Args["path"])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	buffer := make([]byte, bufferSize)
	for {
		bytesread, err := file.Read(buffer)
		if err != nil {
			if err != io.EOF {
				logrus.Error(err)
			}
			break
		}
		c.Plugin.Send(string(buffer[:bytesread]))
	}

	resp.Output = []string{}
	resp.ExitCode = 200

	return nil
}

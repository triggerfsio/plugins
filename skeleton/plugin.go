package main

import (
	"github.com/triggerfsio/plugins"
)

type Command struct {
	Plugin *plugins.Plugin
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

	// do some work here
	c.Plugin.Send("send line to client in realtime...")

	// do some more work
	c.Plugin.Send("notify client again...")

	// and finally set the exitcode and a final message on resp
	resp.ExitCode = 0
	resp.Output = []string{"task has been completed."}

	return nil
}

func (c *Command) Kill(message *plugins.Message, resp *plugins.Response) error {
	// implement your own kill process

	// we will just return here.
	return nil
}

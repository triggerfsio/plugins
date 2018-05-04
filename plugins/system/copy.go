package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/triggerfsio/plugins"
)

type Command struct {
	Plugin *plugins.PluginWrapper
	cmd    *exec.Cmd
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

	if _, ok := message.Args["path"]; !ok {
		c.Plugin.Send("Please specify a path in your plugin arguments.")
		return nil
	}

	path := message.Args["path"]
	if strings.HasSuffix(path, "/") {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			err := os.MkdirAll(path, 0755)
			if err != nil {
				c.Plugin.Send("Couldn't create directory.")
				return err
			}
		}
	} else {
		if _, err := os.Stat(filepath.Dir(path)); os.IsNotExist(err) {
			err := os.MkdirAll(filepath.Dir(path), 0755)
			if err != nil {
				c.Plugin.Send("Couldn't create directory.")
				return err
			}
		}
		file, err := os.Create(path)
		if err != nil {
			c.Plugin.Send("Couldn't create file.")
			return err
		}
		_, err = file.Write([]byte(message.Command[0] + "\n"))
		if err != nil {
			c.Plugin.Send("Couldn't write to file.")
			return err
		}
	}

	resp.Output = []string{"Successfully copied content into file."}
	resp.ExitCode = 200

	return nil
}

func (c *Command) Kill(message *plugins.Message, resp *plugins.Response) error {
	// this plugin has nothing to kill
	return nil
}

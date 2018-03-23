package main

import (
	"bufio"
	"log"
	"os/exec"
	"syscall"

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

	// let's use a function outside of *Command for demonstation purposes.
	err = runCmd(c, message, resp)
	if err != nil {
		return err
	}
	return nil
}

func (c *Command) Kill(message *plugins.Message, resp *plugins.Response) error {
	err := c.cmd.Process.Kill()
	if err != nil {
		return err
	}
	log.Printf("Killed process with pid %d\n", c.cmd.Process.Pid)
	return nil
}

// this function has not the Command receiver by intention.
// just to show an example of how one can use an own function.
func runCmd(c *Command, message *plugins.Message, resp *plugins.Response) error {
	c.cmd = exec.Command("bash", "-c", message.Command)
	stdout, err := c.cmd.StdoutPipe()
	if err != nil {
		log.Printf("Error setting up stdoutpipe: %s", err)
		return err
	}
	stderr, err := c.cmd.StderrPipe()
	if err != nil {
		log.Printf("Error setting up stderrpipe: %s", err)
		return err
	}

	// start the command after having set up the pipe
	if err := c.cmd.Start(); err != nil {
		log.Printf("Error executing the command: %s", err)
		return err
	}
	log.Printf("Started process with pid %d\n", c.cmd.Process.Pid)

	var output []string
	// read command's stdout line by line
	bufout := bufio.NewScanner(stdout)
	buferr := bufio.NewScanner(stderr)

	go func() error {
		for buferr.Scan() {
			line := buferr.Text()
			output = append(output, line)
			c.Plugin.Send(line)
		}
		return nil
	}()

	for bufout.Scan() {
		line := bufout.Text()
		output = append(output, line)
		c.Plugin.Send(line)
	}

	var exitCode int
	if err := c.cmd.Wait(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		} else {
			exitCode = 0
		}
	} else {
		ws := c.cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}

	resp.Output = output
	resp.ExitCode = exitCode

	return nil
}

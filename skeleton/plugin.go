package main

import (
	"github.com/triggerfsio/plugins"
)

// AwesomePlugin is our plugin
// the plugin needs to satisfy the triggerfs plugin interface by implementing its methods.
// there are only two methods which need to be implemented. Init and Kill.
type AwesomePlugin struct {
	// Plugin will hold the triggerfs plugin methods received by plugins.NewPlugin()
	Plugin *plugins.Plugin
}

func main() {
	// init a new triggerfs plugin
	plugin := plugins.NewPlugin()

	// we feed our AwesomePlugin plugin with the newly created plugin
	// and pass it to plugin.Start() to start it.
	// we basically feedback our own plugin back to us (backreference) ready for use.
	plugin.Start(&AwesomePlugin{Plugin: plugin})
}

// Init implements the triggerfs plugin Interface
func (ap *AwesomePlugin) Init(message *plugins.Message, resp *plugins.Response) error {
	// open a channel for realtime communication back to the client.
	err := ap.Plugin.Open(message.Socket)
	if err != nil {
		return err
	}
	// IMPORTANT: remember to defer close the channel or you will get timeouts!
	defer ap.Plugin.Close()

	// now implement your plugin here
	// do some work here
	ap.Plugin.Send("send line to client in realtime...")
	// do some more work
	ap.Plugin.Send("notify client again...")

	// and finally set the exitcode and a final message on resp
	resp.ExitCode = 0
	resp.Output = []string{"task has been completed."}

	return nil
}

// Kill implements the triggerfs plugin Interface
func (ap *AwesomePlugin) Kill(message *plugins.Message, resp *plugins.Response) error {
	// this method will be called by the worker each time we hit a timeout.
	// implement your own kill process here to cleanup things if you need or just return.

	// we will just return here.
	return nil
}

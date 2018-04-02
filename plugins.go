package plugins

import (
	"fmt"
	"log"
	"net/rpc/jsonrpc"
	"time"

	"github.com/natefinch/pie"
	"github.com/pebbe/zmq4"
)

type Message struct {
	Plugin  string            `json:"plugin"`
	Timeout time.Duration     `json:"timeout"`
	Args    map[string]string `json:"args"`
	Command []string          `json:"command"`
	Socket  string            `json:"socket"`
}

type Response struct {
	ExitCode int      `json:"exitcode"`
	Output   []string `json:"output"`
}

type Plugin interface{}

type PluginWrapper struct {
	client *zmq4.Socket
	plugin Plugin
}

func NewPlugin() *PluginWrapper {
	return &PluginWrapper{}
}

func (pl *PluginWrapper) Start(plugin Plugin) {
	log.SetPrefix("[command plugin log] ")

	p := pie.NewProvider()
	if err := p.RegisterName("Plugin", plugin); err != nil {
		log.Fatalf("failed to register Plugin: %s", err)
	}
	stdio, err := zmq4.NewSocket(zmq4.PAIR)
	if err != nil {
		log.Fatalf("failed to initialize socket: %s", err)
	}
	pl.client = stdio
	p.ServeCodec(jsonrpc.NewServerCodec)
}

func (pl *PluginWrapper) Send(data string) error {
	_, err := pl.client.SendMessage(data)
	if err != nil {
		return err
	}
	return nil
}

func (pl *PluginWrapper) Close() {
	if err := recover(); err != nil {
		pl.Send(fmt.Sprintf("Fatal error (panic) in plugin: %s", err))
		log.Printf("Fatal error (panic): %s", err)
	}
	pl.Send("CLOSE")
	pl.client.Close()
}

func (pl *PluginWrapper) Open(socket string) error {
	err := pl.client.Connect(socket)
	if err != nil {
		return err
	}
	return nil
}

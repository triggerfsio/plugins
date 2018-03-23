package plugins

import (
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
	Command string            `json:"command"`
	Socket  string            `json:"socket"`
}

type Response struct {
	ExitCode int      `json:"exitcode"`
	Output   []string `json:"output"`
}

type Pluginer interface {
	Init(*Message, *Response) error
	Kill(*Message, *Response) error
}

type Plugin struct {
	client *zmq4.Socket
}

type PluginWrapper struct {
	plugin Pluginer
	*Plugin
}

func NewPlugin() *Plugin {
	return &Plugin{}
}

func (pl *Plugin) Start(plugin Pluginer) {
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

func (pl *Plugin) Send(data string) error {
	_, err := pl.client.SendMessage(data)
	if err != nil {
		return err
	}
	return nil
}

func (pl *Plugin) Close() {
	pl.client.SendMessage("CLOSE")
	pl.client.Close()
}

func (pl *Plugin) Open(socket string) error {
	err := pl.client.Connect(socket)
	if err != nil {
		return err
	}
	return nil
}

func (pw *PluginWrapper) Init(message *Message, resp *Response) error {
	if message == nil {
		log.Fatalln("No message received. Aborting.")
	}
	err := pw.plugin.Init(message, resp)
	if err != nil {
		return err
	}
	return nil
}

func (pw *PluginWrapper) Kill(message *Message, resp *Response) error {
	err := pw.plugin.Kill(message, resp)
	if err != nil {
		return err
	}
	return nil
}

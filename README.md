# Official TriggerFS Plugins Repository (https://triggerfs.io)

## Core plugins of triggerFS
### triggerFS is a distributed, realtime message passing and trigger system.

#### Description
This is the plugins repository with the core plugins created and maintained by triggerFS.

This repository serves as both the lib being used for writing your own plugins (plugins.go) and the reposiroty for the core plugins in the plugins folder.

The plugins are well tested and known to be working. If you experience any problems please open an issue so we can help you right away.

There is a plugin skeleton for easy copy pasting boilerplate which serves you with a clean plugin body.
https://github.com/triggerfsio/plugins/blob/master/skeleton/plugin.go

You can use it to write your own plugins.

#### Usage
Clone this repository using `go get`:
```
go get github.com/triggerfsio/plugins
```
It should be now available in your $GOPATH/src/github.com/triggerfsio/ folder.

If you haven't done so, set the path of the plugins directory in your toml configuration file that is used by your triggerFS workers:
```
...
### WORKERS SECTION
[workers]
# path to plugins folder
pluginspath = "/home/$yourusername/gocode/src/github.com/triggerfsio/plugins/plugins"
...
```
**Notice**: Remember to use absolute paths in your toml file.

Go to the folder of your desired plugin and build it with `go build`:
```
cd $GOPATH/src/github.com/triggerfsio/plugins/plugins/command
go build command.go
```

You will now have a binary called command in that directory and it will be ready for use.
Go use it with your triggerfs-client. Read more about it in the triggerFS docs at https://triggerfs.io.

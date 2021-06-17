package ctl

import (
	"log"
	"os"

	"github.com/WangYiwei-oss/Docker160/commands"
	"github.com/WangYiwei-oss/cli"
)

func init() {
	//先将log的输出定为标准输出，方便调试，后面改为文件
	log.SetOutput(os.Stdout)
}

var app *cli.App

func NewCliApp() *cli.App {
	if app == nil {
		app = cli.NewApp()
		app.Name = "160-Docker"
		app.Usage = "练习用简易版Docker，出处为<<自己动手写Docker>>"
	}
	cmds := make([]commands.Command, 0)
	cmds = append(cmds, commands.NewRunCommand())
	cmds = append(cmds, commands.NewLogCommand())
	cmds = append(cmds, commands.NewInitCommand())
	cmds = append(cmds, commands.NewPsCommand())
	cmds = append(cmds, commands.NewExecCommand())
	cmds = append(cmds, commands.NewStopCommand())
	cmds = append(cmds, commands.NewRmCommand())
	cmds = append(cmds, commands.NewCommitCommand())
	for _, cmd := range cmds {
		addCommand(cmd)
	}
	return app
}

func addCommand(c commands.Command) {
	app.Commands = append(app.Commands, c.GetCliCommand())
}

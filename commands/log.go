package commands

import (
	"log"

	"github.com/WangYiwei-oss/cli"
)

type LogCommand struct {
	LogCommand *cli.Command
}

func NewLogCommand() *LogCommand {
	logcmd := &cli.Command{
		Name:  "log",
		Usage: "查看日志",
		Action: func(c *cli.Context) error {
			log.Println("执行了log")
			return nil
		},
	}
	return &LogCommand{
		LogCommand: logcmd,
	}
}

func (l *LogCommand) GetCliCommand() *cli.Command {
	return l.LogCommand
}

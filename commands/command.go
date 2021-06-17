package commands

import (
	"github.com/WangYiwei-oss/cli"
)

type Command interface {
	GetCliCommand() *cli.Command
}

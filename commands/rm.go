package commands

import (
	"fmt"
	"log"
	"os"

	"../container"
	"github.com/WangYiwei-oss/cli"
)

type RmCommand struct {
	RmCommand *cli.Command
}

func NewRmCommand() *RmCommand {
	rmcmd := &cli.Command{
		Name:  "rm",
		Usage: "进入容器",
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf("缺少参数")
			}
			containerName := c.Args().Get(0)
			removeContainer(containerName)
			return nil
		},
	}
	return &RmCommand{
		RmCommand: rmcmd,
	}
}

func (r *RmCommand) GetCliCommand() *cli.Command {
	return r.RmCommand
}

func removeContainer(containerName string) {
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		log.Printf("get container info failed %s\n", err)
		return
	}
	if containerInfo.Status != container.STOP {
		log.Printf("不能停止running容器\n")
		return
	}
	DestoryWorkSpace(containerName, containerInfo.Volume)
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err := os.RemoveAll(dirUrl); err != nil {
		log.Printf("remove container info failed: %s", err)
		return
	}
	log.Printf("提示：已删除%s\n", containerInfo.Id)
}

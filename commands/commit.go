package commands

import (
	"fmt"
	"os/exec"

	"github.com/WangYiwei-oss/cli"
)

type CommitCommand struct {
	CommitCommand *cli.Command
}

func NewCommitCommand() *CommitCommand {
	commitcmd := &cli.Command{
		Name:  "commit",
		Usage: "打包容器",
		Action: func(c *cli.Context) error {
			if c.NArg() < 2 {
				return fmt.Errorf("缺失参数")
			}
			containerName := c.Args().Get(0)
			imageName := c.Args().Get(1)
			commitContainer(containerName, imageName)
			return nil
		},
	}
	return &CommitCommand{
		CommitCommand: commitcmd,
	}
}

func (c *CommitCommand) GetCliCommand() *cli.Command {
	return c.CommitCommand
}

func commitContainer(containerName, imageName string) error {
	mntURL := fmt.Sprintf(MntUrl, containerName)
	mntURL += "/"
	imageTar := RootUrl + "/" + imageName + ".tar"
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		return err
	}
	return nil
}

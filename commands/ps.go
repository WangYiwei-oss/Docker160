package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"text/tabwriter"

	"../container"
	"github.com/WangYiwei-oss/cli"
)

type PsCommand struct {
	PsCommand *cli.Command
}

func NewPsCommand() *PsCommand {
	pscmd := &cli.Command{
		Name:  "ps",
		Usage: "查看容器信息",
		Action: func(c *cli.Context) error {
			ListContainers()
			return nil
		},
	}
	return &PsCommand{
		PsCommand: pscmd,
	}
}
func ListContainers() {
	//从文件中获取所有的容器信息
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, "")
	dirUrl = dirUrl[:len(dirUrl)-1]
	files, _ := ioutil.ReadDir(dirUrl)
	var containers []*container.ContainerInfo
	for _, file := range files {
		tmpContainer, err := getContainerInfo(file)
		if err != nil {
			log.Println("读取容器信息失败", err)
			continue
		}
		containers = append(containers, tmpContainer)
	}
	//格式化打印输出相关代码
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, item := range containers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreateTime)
	}
	w.Flush()
}
func getContainerInfo(file os.FileInfo) (*container.ContainerInfo, error) {
	containerName := file.Name()
	configFileDir := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFileDir = configFileDir + container.ConfigName
	content, err := ioutil.ReadFile(configFileDir)
	if err != nil {
		return nil, err
	}
	var containerInfo container.ContainerInfo
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		return nil, err
	}
	return &containerInfo, nil
}
func (p *PsCommand) GetCliCommand() *cli.Command {
	return p.PsCommand
}

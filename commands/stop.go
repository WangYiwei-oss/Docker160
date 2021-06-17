package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"syscall"

	"../container"
	"github.com/WangYiwei-oss/cli"
)

type StopCommand struct {
	StopCommand *cli.Command
}

func NewStopCommand() *StopCommand {
	stopcmd := &cli.Command{
		Name:  "stop",
		Usage: "停止容器",
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf("缺少参数")
			}
			containerName := c.Args().Get(0)
			stopContainer(containerName)
			return nil
		},
	}
	return &StopCommand{
		StopCommand: stopcmd,
	}
}

func (s *StopCommand) GetCliCommand() *cli.Command {
	return s.StopCommand
}

func getContainerInfoByName(containerName string) (*container.ContainerInfo, error) {
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirUrl + container.ConfigName
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}
	var containerInfo container.ContainerInfo
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return nil, err
	}
	return &containerInfo, nil
}

func stopContainer(containerName string) {
	//根据容器名获取对应的主进程PID
	pid, err := getContainerPidByName(containerName)
	if err != nil {
		log.Printf("Get container pid failed%s", err)
		return
	}
	pidInt, _ := strconv.Atoi(pid)
	//系统调用kill可以发送信号给进程，通过传递syscall.SIGTERM信号去杀掉容器进程
	if err := syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		return
	}
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		log.Printf("Get container info failed %s", err)
		return
	}
	containerInfo.Status = container.STOP
	containerInfo.Pid = " "
	newContentBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Printf("Json Marshal error%s", err)
		return
	}
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirUrl + container.ConfigName
	if err := ioutil.WriteFile(configFilePath, newContentBytes, 0622); err != nil {
		log.Printf("Write Config File Failed:%s", err)
	}
}

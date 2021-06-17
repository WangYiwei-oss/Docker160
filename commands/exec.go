package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"../container"
	_ "../nsenter"
	"github.com/WangYiwei-oss/cli"
)

type ExecCommand struct {
	ExecCommand *cli.Command
}

const ENV_EXEC_PID = "mydocker_pid"
const ENV_EXEC_CMD = "mydocker_cmd"

func NewExecCommand() *ExecCommand {
	execcmd := &cli.Command{
		Name:  "exec",
		Usage: "进入容器",
		Action: func(c *cli.Context) error {
			//for callback,就是说如果有这个环境变量的话，那说明已经调用了CGo的代码了，那就不需要再继续接下来的代码了，否则就会无限套娃
			if os.Getenv(ENV_EXEC_PID) != "" {
				log.Printf("pid callback pid %v\n", os.Getpid())
				return nil
			}
			//我们希望的格式是mydocker exec 容器名 命令
			if c.NArg() < 2 {
				return fmt.Errorf("Missing container name or command")
			}
			containerName := c.Args().Get(0)
			var commandArray []string
			for _, arg := range c.Args().Tail() {
				commandArray = append(commandArray, arg)
			}
			ExecContainer(containerName, commandArray)
			return nil
		},
	}
	return &ExecCommand{
		ExecCommand: execcmd,
	}
}

//这里是根据容器名字获取对应容器的PID
func getContainerPidByName(containerName string) (string, error) {
	//拼接正确的路径
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirUrl + container.ConfigName
	//获取文件内容
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return "", err
	}
	var containerInfo container.ContainerInfo
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return "", err
	}
	return containerInfo.Pid, nil
}

func ExecContainer(containerName string, comArray []string) {
	pid, err := getContainerPidByName(containerName)
	if err != nil {
		log.Printf("根据容器名称获取pid错误%v\n", err)
		return
	}
	cmdStr := strings.Join(comArray, " ")
	//这里是重点
	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	os.Setenv(ENV_EXEC_PID, pid)
	os.Setenv(ENV_EXEC_CMD, cmdStr)
	//获取对应PID的环境变量，其实也就是容器的环境变量
	containerEnvs := getEnvByPid(pid)
	//将宿主机的环境变量和容器的环境变量都放置到exec进程内
	cmd.Env = append(os.Environ(), containerEnvs...)
	if err := cmd.Run(); err != nil {
		log.Printf("Exec container %s error %v", containerName, err)
	}
}

func getEnvByPid(pid string) []string {
	//进程环境变量存放的位置是/proc/{pid}/environ
	path := fmt.Sprintf("/proc/%s/environ", pid)
	contentBytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("读取文件%s失败%s", path, err)
		return nil
	}
	//文件中环境变量分隔符为\u0000
	envs := strings.Split(string(contentBytes), "\u0000")
	return envs

}

func (e *ExecCommand) GetCliCommand() *cli.Command {
	return e.ExecCommand
}

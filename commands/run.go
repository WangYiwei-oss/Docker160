package commands

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"../cgroups"
	"../container"
	"github.com/WangYiwei-oss/cli"
)

var (
	RootUrl       string = "/home/wangyiwei/mydocker-test"
	MntUrl        string = "/home/wangyiwei/mydocker-test/mnt/%s"
	WriteLayerUrl string = "/home/wangyiwei/mydocker-test/writeLayer/%s"
)

type RunCommand struct {
	RunCommand *cli.Command
}

func NewRunCommand() *RunCommand {
	runcmd := &cli.Command{
		Name:  "run",
		Usage: "运行容器",
		//这里注册了两个flage,就是比如Docker run -ti /bin/sh这里的ti
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "ti",
				Usage: "前台运行",
			},
			&cli.StringFlag{
				Name:  "m",
				Usage: "限制内存",
			},
			&cli.StringFlag{
				Name:  "v",
				Usage: "挂载数据卷，形如/src:/des",
			},
			&cli.BoolFlag{
				Name:  "d",
				Usage: "后台运行",
			},
			&cli.StringFlag{
				Name:  "name",
				Usage: "指定容器名字",
			},
			&cli.StringSliceFlag{
				Name:  "e",
				Usage: "设置环境变量",
			},
		},
		Action: func(c *cli.Context) error {
			//现在开始写run的逻辑，也就是Docker run /bin/sh执行的东西
			//Docker run /bin/sh是这样子执行的：将此进程fork出来一个新的进程，fork过程中要对NameSpace进行隔离。然后执行容器的初始化过程，包括指定程序/bin/sh的运行
			if c.NArg() < 1 {
				log.Fatal("参数缺失")
			}
			createTty := c.Bool("ti")
			detachContainer := c.Bool("d")
			if createTty && detachContainer {
				log.Fatal("不能同时后台和前台运行")
			}
			rawvolume := c.String("v")
			//第一个参数标记了是否是前台运行，传参：第二个参数负责传递参数
			resource := &cgroups.ResourceConfig{
				Memory: c.String("m"),
				Cpuset: c.String("cpuset"),
			}
			containerName := c.String("name")
			cmdArray := c.Args().Slice()
			imageName := cmdArray[0]
			envSlice := c.StringSlice("e")
			run(createTty, cmdArray[1:], resource, rawvolume, containerName, imageName, envSlice)
			return nil
		},
		After: func(c *cli.Context) error {
			return nil
		},
	}
	return &RunCommand{
		RunCommand: runcmd,
	}
}

func (r *RunCommand) GetCliCommand() *cli.Command {
	return r.RunCommand
}

func newParentProcess(tty bool, containerName, volume, imageName string, envSlice []string) (*exec.Cmd, *os.File) {
	//传参：为了传递参数，使用了匿名管道
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		log.Fatal("传参管道初始化错误", err)
	}
	//实际上就是在执行Docker init /bin/sh
	cmd := exec.Command("/proc/self/exe", "init", containerName)
	//设置NameSpace的隔离
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWIPC,
	}
	//传参：将读管道作为额外的文件句柄传递给cmd，也就是重新执行的本程序
	cmd.ExtraFiles = []*os.File{readPipe}
	cmd.Env = append(os.Environ(), envSlice...)
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	//让子进程的工作目录处于挂载目录下
	cmd.Dir = fmt.Sprintf(MntUrl, containerName)
	return cmd, writePipe
}

func run(tty bool, commands []string, r *cgroups.ResourceConfig, rawvolume string, containerName string, imageName string, envSlice []string) {
	//挂载aufs文件系统,以及volume
	NewWorkSpace(containerName, rawvolume, imageName)
	//newParentProcess函数实现了NameSpace的隔离
	parent, writePipe := newParentProcess(tty, containerName, rawvolume, imageName, envSlice)
	//传参：将参数写入传递参数用的匿名管道
	writePipe.WriteString(strings.Join(commands, " "))
	writePipe.Close()
	if err := parent.Start(); err != nil {
		log.Fatal("Start myself failed", err)
	}
	//记录容器信息
	_, err := recordContainerInfo(parent.Process.Pid, commands, containerName, rawvolume)
	if err != nil {
		log.Println("记录容器信息失败", err)
	}
	if r.Memory == "" {
		log.Println("提示：未设置内存资源限制")
	} else {
		cgroupManager := cgroups.NewCgroupManager("160Docker")
		defer cgroupManager.Remove()
		cgroupManager.Resource = *r
		cgroupManager.Set()
		cgroupManager.Apply(parent.Process.Pid)
	}
	if tty {
		parent.Wait()
		DestoryWorkSpace(containerName, rawvolume)
		deleteContainerInfo(containerName)
	}
}

//记录容器信息
func recordContainerInfo(pid int, cmdArray []string, containerName string, volume string) (string, error) {
	id := container.RandStringBytes(10)
	createTime := time.Now().Format("2006-01-02 15:04:05")
	if containerName == "" {
		containerName = id
	}
	commands := strings.Join(cmdArray, " ")
	containerInfo := &container.ContainerInfo{
		Id:         id,
		Pid:        strconv.Itoa(pid),
		Command:    commands,
		CreateTime: createTime,
		Name:       containerName,
		Status:     container.RUNNING,
		Volume:     volume,
	}
	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Printf("json化容器信息错误%v\n", err)
	}
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err := os.MkdirAll(dirUrl, 0622); err != nil {
		log.Printf("创建文件夹出错%v,%v\n", dirUrl, err)
		return "", nil
	}
	fileName := dirUrl + "/" + container.ConfigName
	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		log.Printf("创建文件夹出错%v,%v\n", fileName, err)
		return "", nil
	}
	file.WriteString(string(jsonBytes))
	return containerName, nil
}
func deleteContainerInfo(containerId string) {
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	if err := os.RemoveAll(dirUrl); err != nil {
		log.Println("删除容器信息失败")
	}
}

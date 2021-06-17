package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"

	"github.com/WangYiwei-oss/cli"
)

type InitCommand struct {
	InitCommand *cli.Command
}

func NewInitCommand() *InitCommand {
	initcmd := &cli.Command{
		Name:  "init",
		Usage: "内部调用，执行初始化工作",
		Before: func(c *cli.Context) error {
			runInitialization()
			return nil
		},
		Action: func(c *cli.Context) error {
			runContainerInitProcess()
			return nil
		},
		After: func(c *cli.Context) error {
			return nil
		},
	}
	return &InitCommand{
		InitCommand: initcmd,
	}
}

func pivotRoot(root string) error {
	//为了使当前root和老root不在同一文件系统下，需要重新挂载一遍
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("Mount rootfs to itself error %v\n", err)
	}
	pivotDir := path.Join(root, "old_root")
	//创建rootfs/old_root文件来储存老的root
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return err
	}
	//将父root设为private
	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	//系统调用pivotroot,挂载新的root，并将老root放到old_root文件中
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root failed %v\n", err)
	}
	syscall.Chdir("/")
	pivotDir = path.Join("/", "old_root")
	//这里我理解的root文件系统就像是linux刚刚启动时，就会自动挂载一个默认的linux的文件系统，有了这个文件系统，我们的操作系统才能够启动。这里搞了一个新的文件系统，所以此时"/"里面可以理解为一个新的操作系统所需的文件系统。而/old_root文件夹挂载的是本来linux的默认文件系统
	//取消挂载老的root
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount old_root fs failed%v\n", err)
	}
	//删除临时文件
	return os.Remove(pivotDir)
}

func runInitialization() {
	//进行一些挂载的操作，将proc挂载过来，方便使用ps等命令

	pwd, err := os.Getwd()
	if err != nil {
		log.Println("Get current location failed")
	}
	err = pivotRoot(pwd)
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	//挂载虚存
	syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_STRICTATIME, "mode=755")
}

func (i *InitCommand) GetCliCommand() *cli.Command {
	return i.InitCommand
}

func runContainerInitProcess() error {
	//传参：从管道中读取命令
	readPipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(readPipe)
	if err != nil {
		log.Fatal("初始化容器错误，读取管道命令错误", err)
	}
	cmdArray := strings.Split(string(msg), " ")
	//下面这句是自动补全,比如我输入了sh就可以自动查找到/bin/sh
	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		log.Fatal("LookPath Failed", err)
	}
	if err := syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil {
		log.Fatal("执行指定程序出错", err)
	}
	return nil
}

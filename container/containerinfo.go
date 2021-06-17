package container

import (
	"math/rand"
	"time"
)

type ContainerInfo struct {
	Pid        string //容器运行的pid
	Id         string //随机生成的id
	Name       string //容器的名字
	Command    string //容器运行的命令
	CreateTime string //创建时间
	Status     string //状态
	Volume     string //挂载信息
}

//状态的枚举
var (
	RUNNING             string = "running"
	STOP                string = "stopped"
	Exit                string = "exited"
	DefaultInfoLocation string = "/var/run/mydocker/%s/"
	ConfigName          string = "config.json"
)

//随机生成id的函数
func RandStringBytes(n int) string {
	letterBytes := "1234567890"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

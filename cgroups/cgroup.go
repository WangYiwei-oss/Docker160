package cgroups

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

//由于有多种资源的限制，比如内存，cpu，所以这里将他定义为接口
type Subsystem interface {
	Name() string
	//设置某个cgroup在这个Subsystem中的资源限制,也就是在代表cgroup的文件夹内将资源限制写入代表subsystem的文件
	Set(path string, r *ResourceConfig) error
	//将进程添加进代表cgroup的文件夹内的tasks文件中
	Apply(path string, pid int) error
	Remove(path string) error
}

//取得当前subsystem在虚拟文件系统中的路径
func FindCgroupMountpoint(subsystem string) string {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		log.Printf("Warning：资源限制：打开/proc/self/mountinfo失败:%v\n", err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		txt := strings.Split(scanner.Text(), " ")
		if txt[len(txt)-1] == "rw,"+subsystem {
			return txt[4]
		}
	}
	return ""
}

func GetCgroupPath(subsystem string, cgroupPath string, autoCreate bool) string {
	cgroupRoot := FindCgroupMountpoint(subsystem)
	if _, err := os.Stat(path.Join(cgroupRoot, cgroupPath)); err == nil || (autoCreate && os.IsNotExist(err)) {
		if os.IsNotExist(err) {
			if err := os.Mkdir(path.Join(cgroupRoot, cgroupPath), 0755); err != nil {
				return ""
			}
		}
		//创建成功或者已经存在就返回其路径
		return path.Join(cgroupRoot, cgroupPath)
	} else {
		return ""
	}
}

//内存限制的Subsystem
type MemorySubsystem struct{}

//返回subsystem的名字，也就是memory
func (m *MemorySubsystem) Name() string {
	return "memory"
}
func (m *MemorySubsystem) Set(filepath string, r *ResourceConfig) error {
	cgroupPath := GetCgroupPath(m.Name(), filepath, true)
	if r.Memory == "" || cgroupPath == "" {
		log.Println("内存资源限制未配置或路径错误")
	} else {
		if err := ioutil.WriteFile(path.Join(cgroupPath, "memory.limit_in_bytes"), []byte(r.Memory), 0644); err != nil {
			return fmt.Errorf("write to memory failed %v", err)
		} else {
		}
	}
	return nil
}
func (m *MemorySubsystem) Apply(filepath string, pid int) error {
	cgroupPath := GetCgroupPath(m.Name(), filepath, true)
	if cgroupPath == "" {
		return fmt.Errorf("寻找CgroupPath失败")
	}
	if err := ioutil.WriteFile(path.Join(cgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("set cgroup proc fail %v", err)
	}
	return nil
}

func (m *MemorySubsystem) Remove(filepath string) error {
	cgroupPath := GetCgroupPath(m.Name(), filepath, false)
	if cgroupPath == "" {
		return fmt.Errorf("Remove cgroup failed \n")
	} else {
		return os.RemoveAll(cgroupPath)
	}
}

//继续写限制cpu数量的
type CpuSetSubsystem struct{}

//返回subsystem的名字，也就是memory
func (m *CpuSetSubsystem) Name() string {
	return "cpuset"
}
func (m *CpuSetSubsystem) Set(filepath string, r *ResourceConfig) error {
	cgroupPath := GetCgroupPath(m.Name(), filepath, true)
	if r.Memory == "" || cgroupPath == "" {
		log.Println("cpuset资源限制未配置或路径错误")
	} else {
		if err := ioutil.WriteFile(path.Join(cgroupPath, "cpuset.cpus"), []byte(r.Memory), 0644); err != nil {
			return fmt.Errorf("write to memory failed %v", err)
		} else {
			log.Println("cpuset资源限制写入成功", path.Join(cgroupPath, "memory.limit_in_bytes"), r.Memory)
		}
	}
	return nil
}
func (m *CpuSetSubsystem) Apply(filepath string, pid int) error {
	cgroupPath := GetCgroupPath(m.Name(), filepath, true)
	if cgroupPath == "" {
		return fmt.Errorf("寻找CgroupPath失败")
	}
	if err := ioutil.WriteFile(path.Join(cgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("set cgroup proc fail %v", err)
	}
	log.Println("pid写入tasks")
	return nil
}
func (m *CpuSetSubsystem) Remove(filepath string) error {
	cgroupPath := GetCgroupPath(m.Name(), filepath, false)
	if cgroupPath == "" {
		return fmt.Errorf("Remove cgroup failed \n")
	} else {
		return os.RemoveAll(cgroupPath)
	}
}

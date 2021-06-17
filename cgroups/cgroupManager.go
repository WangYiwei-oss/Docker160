package cgroups

import (
	"fmt"
	"log"
)

type CgroupManager struct {
	Path     string
	Resource ResourceConfig
}

var (
	SubsystemInfo = []Subsystem{
		&MemorySubsystem{},
	}
)

func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		Path: path,
	}
}
func (c *CgroupManager) SetResource(r *ResourceConfig) {
	c.Resource = *r
}
func (c *CgroupManager) Apply(pid int) {
	for _, subSys := range SubsystemInfo {
		err := subSys.Apply(c.Path, pid)
		if err != nil {
			log.Printf("Apply %v 错误:%v\n", subSys.Name(), err)
		}
	}
}

func (c *CgroupManager) Set() {
	for _, subSys := range SubsystemInfo {
		err := subSys.Set(c.Path, &c.Resource)
		if err != nil {
			fmt.Printf("Set %v 错误:%v\n", subSys.Name(), err)
		}
	}
}

func (c *CgroupManager) Remove() {
	for _, subSys := range SubsystemInfo {
		err := subSys.Remove(c.Path)
		if err != nil {
			fmt.Printf("Remove %v 错误:%v\n", subSys.Name(), err)
		}
	}
}

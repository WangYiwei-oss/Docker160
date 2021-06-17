package cgroups

//资源限制配置的结构体，暂时先只写内存和cpu
type ResourceConfig struct {
	Memory string
	Cpuset string
}

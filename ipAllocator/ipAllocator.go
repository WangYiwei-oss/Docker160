package ipAllocator

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"path"
	"strings"
)

const ipamDefaultAllocatorPath = "/var/run/mydocker/network/ipam/subnet.json"

type IPAM struct {
	//分配文件存放位置
	SubnetAllocatorPath string
	//网段和位图算法的数组map,key是网段，value是分配的位图数组
	Subnets *map[string]string
}

var IpAllocator = &IPAM{
	SubnetAllocatorPath: ipamDefaultAllocatorPath,
}

//加载网段地址分配信息
func (ipam *IPAM) load() error {
	if _, err := os.Stat(ipam.SubnetAllocatorPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}
	subnetConfigFile, err := os.Open(ipam.SubnetAllocatorPath)
	defer subnetConfigFile.Close()
	if err != nil {
		return err
	}
	subnetJson := make([]byte, 2000)
	n, err := subnetConfigFile.Read(subnetJson)
	if err != nil {
		return err
	}
	err = json.Unmarshal(subnetJson[:n], ipam.Subnets)
	if err != nil {
		log.Printf("Error dump allocation info%v\n", err)
		return err
	}
	return nil
}

//储存网段地址分配信息
func (ipam *IPAM) dump() error {
	ipamConfigFileDir, _ := path.Split(ipam.SubnetAllocatorPath)
	if _, err := os.Stat(ipamConfigFileDir); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(ipamConfigFileDir, 0644)
		} else {
			return err
		}
	}
	//打开储存文件，os.O_TRUNC表示如果存在则清空,os.O_CREATE表示如果不存在则创建
	subnetConfigFile, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	defer subnetConfigFile.Close()
	if err != nil {
		return err
	}
	//序列化ipam对象到json串
	ipamConfigJson, err := json.Marshal(ipam.Subnets)
	if err != nil {
		return err
	}
	_, err = subnetConfigFile.Write(ipamConfigJson)
	if err != nil {
		return err
	}
	return nil
}

//在网段中分配一个可用的IP地址
func (ipam *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	//存放网段中地址分配信息的数组
	ipam.Subnets = &map[string]string{}
	//从文件中加载已经分配的网段信息
	err = ipam.load()
	if err != nil {
		log.Printf("Error load allocation info, %v\n", err)
	}
	//net.IPNet.Mask.Size()会返回网段的子网掩码的总长度和网段前面的固定位的长度
	//比如127.0.0.1/8网段的子网掩码是255.0.0.0
	//那么subnet.Mask.Size()的返回值就是前面255所对应的位数和总位数，即8和24
	one, size := subnet.Mask.Size()
	//如果之前没有分配过这个网段，则初始化网段的分配配置
	if _, exist := (*ipam.Subnets)[subnet.String()]; !exist {
		//用0填满这个网段的配置,1<<uint8(size-one)表示这个网段中有多少可用IP数
		//size-one是子网掩码后面的网络位数，2^(size-one)表示网段中的可用IP数
		//而2^(size-one)等价于1<<uint8(size-one)
		(*ipam.Subnets)[subnet.String()] = strings.Repeat("0", 1<<uint8(size-one))
	}
	//遍历网段的位图数组
	for c := range (*ipam.Subnets)[subnet.String()] {
		if (*ipam.Subnets)[subnet.String()][c] == '0' {
			//设置这个为0的序号为1,即分配这个IP
			ipalloc := []byte((*ipam.Subnets)[subnet.String()])
			//go中的string是不能修改的，所以转化成[]byte修改了之后再转回来
			ipalloc[c] = '1'
			(*ipam.Subnets)[subnet.String()] = string(ipalloc)
			//这个ip为初始ip，比如网段192.168.0.0/16，这里就是192.168.0.0
			ip = subnet.IP
			/*
				通过网段的IP与上面的偏移相加计算出分配IP地址，由于IP地址是uint的一个数组，需要通过数组中的每一项加所需的值，比如网段是172.16.0.0/12，数组序号是6555,那么在[172,16,0,0]上依次加[uint8(6555>>24)、uint8(6555>>16)、uint8(6555>>8)、uint8(6555>>0)]，即得到[0,1,0,19]，那么获得的IP就是172.17.0.19
			*/
			for t := uint(4); t > 0; t -= 1 {
				[]byte(ip)[4-t] += uint8(c >> ((t - 1) * 8))
			}
			//由于此处IP是从1开始分配的，所以最后再加1,最终得到分配的IP是172.17.0.20
			ip[3] += 1
			break
		}
	}
	//保存到文件中
	err = ipam.dump()
	if err != nil {
		log.Printf("dump错误%v\n", err)
	}
	return
}

//地址释放的实现
func (ipam *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) error {
	ipam.Subnets = &map[string]string{}
	//从文件中加载网段的分配信息
	err := ipam.load()
	if err != nil {
		log.Printf("Error load allocation info, %v\n", err)
	}
	//计算IP地址在网段位图数组中的索引位置
	c := 0
	releaseIP := ipaddr.To4()
	//由于地址是从1开始分配的，所以转换成索引应-1
	releaseIP[3] -= 1
	for t := uint(4); t > 0; t -= 1 {
		//与分配IP相反，释放IP获得索引的方式是IP地址的每一位相减之后分别左移将对应的数值加到索引上
		c += int(releaseIP[t-1]-subnet.IP[t-1]) << ((4 - t) * 8)
	}
	//将分配的位图数组中索引位置的值置为0
	ipalloc := []byte((*ipam.Subnets)[subnet.String()])
	ipalloc[c] = '0'
	(*ipam.Subnets)[subnet.String()] = string(ipalloc)
	//保存释放掉ip之后的网段IP分配信息
	ipam.dump()
	return nil
}

package network

import (
	"fmt"
	"log"
	"net"
	"strings"

	"../ipAllocator"
	"github.com/vishvananda/netlink"
)

//网络，一个网络包含很多端点
type Network struct {
	Name    string     //网络名
	IpRange *net.IPNet //地址段
	Driver  string     //网络驱动名
}

//网络端点
type Endpoint struct {
	ID          string           `json:"id"`
	Device      netlink.Veth     `json:"dev"`
	IPAddress   net.IP           `json:"ip"`
	MacAddress  net.HardwareAddr `json:"max"`
	PortMapping []string         `json:"portmapping"`
	Network     *Network
}

//网络驱动
type NetworkDriver interface {
	//驱动名
	Name() string
	//创建网络
	Create(subnet string, name string) (*Network, error)
	//删除网络
	Delete(network Network) error
	//连接容器网络端点到网络
	Connect(network *Network, endpoint *Endpoint) error
	//从网络上移除容器网络端点
	Disconnect(network Network, endpoint *Endpoint) error
}

//创建网络,命令eg：Docker netwwork create --subnet 192.168.0.0/24 --driver bridge testbridgenet
func CreateNetwork(driver, subnet, name string) error {
	//将网段的字符串转换成net.IPNet对象
	_, cidr, _ := net.ParseCIDR(subnet)
	gatewayIp, err := ipAllocator.IpAllocator.Allocate(cidr)
	if err != nil {
		return err
	}
	cidr.IP = gatewayIp
	/*调用指定的网络驱动创建网络，这里的drivers字典是各个网络驱动的实例字典，通过调用网络驱动的Create方法创建网络，目前仅仅实现了Bridge驱动*/
	return nil
}

//实现BridgeNetworkDriver
type BridgeNetworkDriver struct {
}

//subnet类似于192.168.0.0/24
func (d *BridgeNetworkDriver) Create(subnet string, name string) (*Network, error) {
	//ip为192.168.0.1,ipRange为24
	ip, ipRange, _ := net.ParseCIDR(subnet)
	ipRange.IP = ip
	n := &Network{
		Name:    name,
		IpRange: ipRange,
	}
	err := d.initBridge(n)
	if err != nil {
		log.Printf("error init bridge: %v\n", err)
	}
	return n, err
}

//BridgeDrive初始化linux bridge
func (d *BridgeNetworkDriver) initBridge(n *Network) error {
	//1.创建Bridge虚拟设备
	bridgeName := n.Name
	if err := createBridgeInterface(bridgeName); err != nil {
		return fmt.Errorf("Error add bridge: %s, %v\n", bridgeName, err)
	}
	//2.设置Bridge设备的地址和路由
	gatewayIP := *n.IpRange
	gatewayIP.IP = n.IpRange.IP
	/*
		if err := setInterfaceIP(bridgeName, gatewayIP.String()); err != nil {
			return err
		}*/
	return nil
}

//创建linux bridge设备
func createBridgeInterface(bridgeName string) error {
	//1.先检查是否已经存在同名设备
	_, err := net.InterfaceByName(bridgeName)
	if err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return err
	}
	//初始化一个netlink的link基础对象，Link的名字即Bridge虚拟设备的名字
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName
	br := &netlink.Bridge{LinkAttrs: la}
	if err := netlink.LinkAdd(br); err != nil {
		return fmt.Errorf("Bridge creation filed for bridge: %v\n", err)
	}
	return nil
}

//设置Bridge设备的地址和路由,例如setInterfaceIP("testBridge","192.168.0.1/24")
/*
func setInterfaceIP(name string, rawIP string) error {
	//通过netlink的LinkByName方法找到需要设置的网络端口
	iface, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}
		//由于netlink.ParseIPNet是对net.ParseCIDR的一个封装，因此可以将net.ParseCIDR的返回值中的IP和net整合

	ipNet, err := netlink.ParseIPNet(rawIP)
	if err != nil {
		return err
	}
	//通过netlink.AddrAdd给网络接口配置地址，相当于ip addr add xxx命令
	//同时如果配置了地址所在网段的信息，例如192.168.0.0/24，还会配置路由表192.168.0.0/24转发到这个testBridge的网络接口上
	return netlink.AddrAdd(iface, addr)
}
*/

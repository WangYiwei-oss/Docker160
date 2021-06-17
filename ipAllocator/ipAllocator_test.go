package ipAllocator

import (
	"fmt"
	"net"
	"testing"
)

func Test1(t *testing.T) {
	_, ipnet, _ := net.ParseCIDR("192.168.0.67/31")
	ip, _ := IpAllocator.Allocate(ipnet)
	fmt.Printf("%v\n", ip)
}

/*
func Test2(t *testing.T) {
	ip, ipnet, _ := net.ParseCIDR("192.168.0.2/24")
	ipAllocator.Release(ipnet, &ip)
}
*/

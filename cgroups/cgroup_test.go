package cgroups

import (
	"fmt"
	"testing"
)

func Test1(t *testing.T) {
	fmt.Println(GetCgroupPath("memory", "mydocker", true))
}

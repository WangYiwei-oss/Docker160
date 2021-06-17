package network

import (
	"fmt"
	"testing"
)

func Test1(t *testing.T) {
	err := createBridgeInterface("br0")
	if err != nil {
		fmt.Println(err)
	}
}

package util

import (
	"fmt"
	"net"
)

func GetAvailablePort() (int, error) {
	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:0", "0.0.0.0"))
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		return 0, err
	}

	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil

}
func MacAddr() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		panic("net interfaces" + err.Error())
	}
	mac := ""
	for _, netInterface := range interfaces {
		mac = netInterface.HardwareAddr.String()
		if len(mac) == 0 {
			continue
		}
		break
	}
	return mac
}

package util

import "net"

var IPv4 int8 = 4
var IPv6 int8 = 6
var InvalidIPVersion int8 = 0

func GetCIDRVersion(cidr string) (int8, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return InvalidIPVersion, err
	}
	if ipNet.IP.To4() != nil {
		return IPv4, nil
	}
	return IPv6, nil
}

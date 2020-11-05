package utils

import (
	"net"
	"os"
)

//FileExists function
func FileExists(DbPath string) bool {
	info, err := os.Stat(DbPath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

//IsValidIPV4Address function
func IsValidIPV4Address(ipAddress string) bool {
	addr := net.ParseIP(ipAddress)
	if addr == nil {
		return false
	}
	if addr.To4() == nil {
		return false
	}
	return true
}

//go:build windows

package sysutils

import "os/user"

func GetSessionUser() (*user.User, error) {
	return user.Current()
}

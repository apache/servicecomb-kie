package cipherutil

import (
	"github.com/go-chassis/go-chassis/v2/security/cipher"
	"github.com/go-chassis/openlog"
)

// TryDecrypt return the src when decrypt failed
func TryDecrypt(src string) string {
	res, err := cipher.Decrypt(src)
	if err != nil {
		openlog.Info("cipher fallback: " + err.Error())
		res = src
	}
	return res
}

// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package common

import (
	"os"

	"github.com/kweaver-ai/kweaver-go-lib/crypto"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
)

var aesCipter crypto.Cipher

func init() {
	AESKEY := os.Getenv("AES_KEY")
	if AESKEY == "" {
		logger.Fatal("AES_KEY is empty")
	}

	aesCipter = crypto.NewAESCipher(AESKEY)
}

// 解密
func DecryptPassword(pasword string) string {
	if pasword != "" {
		pasword = aesCipter.Decrypt(pasword)
	}
	return pasword
}

// 加密
func EncryptPassword(password string) string {
	if password != "" {
		password = aesCipter.Encrypt(password)
	}
	return password
}

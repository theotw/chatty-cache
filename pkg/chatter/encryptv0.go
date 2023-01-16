/*
 * Copyright (c) The One True Way 2023. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package chatter

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
)

func makeAesKey(phrase string) []byte {
	hash := sha256.New()

	plainText := []byte(phrase)
	hash.Write(plainText)
	ret := hash.Sum([]byte(""))
	return ret
}
func makeRandom256AesKey() []byte {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		panic("cannot read from random, something is very wrong, we need to panic, possible security issye")
	}
	return key
}
func DoAesCBCEncrypt(src, key []byte) ([]byte, error) {

	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	src = addZeroPadding(src, bs)
	iv := make([]byte, bs)
	_, err = rand.Read(iv)
	if err != nil {
		return nil, err
	}

	if len(src)%bs != 0 {
		return nil, errors.New("pad failure")
	}
	cbcMode := cipher.NewCBCEncrypter(block, iv)
	out := make([]byte, len(src))
	cbcMode.CryptBlocks(out, src)
	ret := make([]byte, len(out)+len(iv))
	copy(ret, iv)
	copy(ret[bs:], out)
	return ret, nil
}

func DoAesCBCDecrypt(src, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	out := make([]byte, len(src)-bs)
	iv := src[:bs]

	if len(src)%bs != 0 {
		return nil, errors.New("not padded properly")
	}

	cbcMode := cipher.NewCBCDecrypter(block, iv)
	cbcMode.CryptBlocks(out, src[bs:])

	out = unPadTheZeros(out)
	return out, nil
}
func addZeroPadding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{0}, padding)
	return append(ciphertext, padtext...)
}

func unPadTheZeros(origData []byte) []byte {
	return bytes.TrimFunc(origData,
		func(r rune) bool {
			return r == rune(0)
		})
}

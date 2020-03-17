package crypt

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"strings"
)

var PublicKey *rsa.PublicKey
var PrivateKey *rsa.PrivateKey

func Md5(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

func PrivateKeyToBytes(priv *rsa.PrivateKey) []byte {
	privBytes := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(priv),
		},
	)
	return privBytes
}

func NormalizeText(s string) string {
	return strings.ReplaceAll(s, "\n", "%%%")
}

func DenormalizeText(s string) string {
	return strings.ReplaceAll(s, "%%%", "\n")
}

func PublicKeyToBytes(pub *rsa.PublicKey) []byte {
	pubANS1, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		fmt.Printf("Error(%v)\n", err)
	}

	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubANS1,
	})
	return pubBytes
}

func BytesToPublicKey(pub []byte) *rsa.PublicKey {
	block, _ := pem.Decode(pub)
	enc := x509.IsEncryptedPEMBlock(block)
	b := block.Bytes
	var err error

	if enc {
		b, err = x509.DecryptPEMBlock(block, nil)
		if err != nil {
			fmt.Printf("Error %v\n", err)
			return nil
		}
	}
	ifc, err := x509.ParsePKIXPublicKey(b)
	if err != nil {
		fmt.Printf("Error %v", err)
		return nil
	}
	key, ok := ifc.(*rsa.PublicKey)
	if !ok {
		return nil
	}
	return key
}

func GenKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil
	}
	publicKey := &privateKey.PublicKey
	return privateKey, publicKey
}

func BytesToPrivateKey(priv []byte) *rsa.PrivateKey {
	block, _ := pem.Decode(priv)
	enc := x509.IsEncryptedPEMBlock(block)
	b := block.Bytes
	var err error
	if enc {
		b, err = x509.DecryptPEMBlock(block, nil)
		if err != nil {
			return nil
		}
	}
	key, err := x509.ParsePKCS1PrivateKey(b)
	if err != nil {
		return nil
	}
	return key
}

func DecryptWithPrivateKey(ciphertext []byte, priv *rsa.PrivateKey) []byte {
	hash := sha256.New()
	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, priv, ciphertext, nil)
	if err != nil {
		return nil
	}
	return plaintext
}

func EncryptWithPublicKey(plaintext []byte, pub *rsa.PublicKey) []byte {
	hash := sha256.New()
	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, pub, plaintext, nil)
	if err != nil {
		return nil
	}
	return ciphertext
}

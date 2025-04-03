package crypter

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"

	"golang.org/x/crypto/bcrypt"
)

// AES-GCM
func Encrypt(data []byte, key []byte) []byte {
    block, _ := aes.NewCipher(key)
    gcm, _ := cipher.NewGCM(block)
    
    nonce := make([]byte, gcm.NonceSize())
    rand.Read(nonce)

    return gcm.Seal(nonce, nonce, data, nil)
}

func Decrypt(data []byte, key []byte) []byte {
    block, _ := aes.NewCipher(key)
    gcm, _ := cipher.NewGCM(block)
    
    nonce := data[:gcm.NonceSize()]
    ciphertext := data[gcm.NonceSize():]
    
    plaintext, _ := gcm.Open(nil, nonce, ciphertext, nil)
    return plaintext
}

// Использование:
// key := make([]byte, 32) // 256-bit key
// rand.Read(key)
// encrypted := encrypt([]byte("secret"), key)
// decrypted := decrypt(encrypted, key)

func HashPassword(password string) (string, error) {
	hash, err:=bcrypt.GenerateFromPassword([]byte(password),bcrypt.MinCost)
	if err!=nil{
		return "", err
	}

	return string(hash), nil
}

func CompareHashPassword(hash, actualPassword string) bool {
	err:=bcrypt.CompareHashAndPassword([]byte(hash),[]byte(actualPassword))
	
	return err==nil
}
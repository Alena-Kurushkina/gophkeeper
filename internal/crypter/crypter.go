package crypter

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"errors"

	"github.com/Alena-Kurushkina/gophkeeper/internal/gopherror"
	"golang.org/x/crypto/bcrypt"
)
var searchErr = errors.New("cipher: message authentication failed")

// Encrypt шифрует данные с помощью AES-GCM, используя ключ, полученный из пароля
func Encrypt(plaintext []byte, password string) ([]byte, error) {
	// Получаем ключ из пароля с помощью SHA-256
	key := sha256.Sum256([]byte(password))

	// Создаем новый блок AES
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	// Создаем GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Создаем nonce из последних 12 байт ключа
	nonce := make([]byte, gcm.NonceSize())
	copy(nonce, key[len(key)-len(nonce):])

	// Шифруем данные
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	return ciphertext, nil
}

// Decrypt расшифровывает данные, зашифрованные с помощью AES-GCM
func Decrypt(ciphertext []byte, password string) ([]byte, error) {
	// Получаем ключ из пароля
	key := sha256.Sum256([]byte(password))

	// Создаем новый блок AES
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	// Создаем GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// создаём вектор инициализации
	nonce := key[len(key)-gcm.NonceSize():]

	// Расшифровываем данные
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		if searchErr.Error() == err.Error(){
			return nil, gopherror.ErrDecryptAuth
		}else{
			return nil, err
		}
	}

	return plaintext, nil
}

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
package crypter

import (
	"testing"

	"github.com/Alena-Kurushkina/gophkeeper/internal/gopherror"
	"github.com/stretchr/testify/assert"
)

func TestEncrypt(t *testing.T) {
	
	tests := []struct {
		name    string
		encrPassword string
		decrPassword string
		data []byte
	}{
		{
			name:    "valid encrypt",
			encrPassword: "123456",
			decrPassword: "123456",
			data: []byte("{'login': 'login1', 'password': '6784676dfg'}"),
		},
		{
			name:    "wrong password",
			encrPassword: "123456",
			decrPassword: "5457647",
			data: []byte("{'login': 'login1', 'password': '6784676dfg'}"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err:=Encrypt(tt.data,tt.encrPassword)
			assert.NoError(t,err)
			decrypted, err:=Decrypt(encrypted, tt.decrPassword)			
			if tt.encrPassword!=tt.decrPassword{
				assert.ErrorIs(t, err, gopherror.ErrDecryptAuth)
			} else {
				assert.NoError(t,err)
				assert.Equal(t, decrypted, tt.data)
			}
		})
	}
}
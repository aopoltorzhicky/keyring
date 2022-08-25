package keyring

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path"
	"sync"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	file = ".local/share/go_keyring/keyring.yaml"
)

var Keys *File

var (
	defaultPassword = []byte("297ynt237b4tv92ng0m>cy8r4unvch3m")
)

// File -
type File struct {
	pathFile string
	key      []byte
	mx       sync.Mutex
}

// Create -
func Create(secret []byte) error {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	if userHomeDir == "/" {
		userHomeDir = "/root/"
	}
	Keys = new(File)

	if len(secret) == 0 {
		Keys.key = defaultPassword
	} else {
		hash := sha256.New()
		if _, err := hash.Write(secret); err != nil {
			return err
		}
		Keys.key = hash.Sum(nil)
	}

	Keys.pathFile = path.Join(userHomeDir, file)
	if _, err := os.Stat(Keys.pathFile); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(path.Dir(Keys.pathFile), os.ModePerm); err != nil {
			return err
		}

		f, err := os.Create(Keys.pathFile)
		if err != nil {
			return err
		}

		keys := make(keys)
		if err := yaml.NewEncoder(f).Encode(&keys); err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
	}

	return nil
}

// Set - add encrypted `password` to keyring for `service` and `user`
func (k *File) Set(service, user, password string) error {
	k.mx.Lock()
	defer k.mx.Unlock()

	f, keys, err := k.read()
	if err != nil {
		return err
	}
	defer f.Close()

	key := k.getKeyName(service, user)
	encoded, err := k.encode(password)
	if err != nil {
		return err
	}
	keys[key] = encoded
	return k.write(f, keys)
}

// Get - receive password from keyring for `service` and `user`
func (k *File) Get(service, user string) (string, error) {
	k.mx.Lock()
	defer k.mx.Unlock()

	f, keys, err := k.read()
	if err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}

	key := k.getKeyName(service, user)
	encoded, ok := keys[key]
	if !ok {
		return "", errors.Errorf("unknown key for service '%s' and user '%s'", service, user)
	}
	return k.decode(encoded)
}

// Delete - removes password from keyring for `service` and `user`
func (k *File) Delete(service, user string) error {
	k.mx.Lock()
	defer k.mx.Unlock()

	f, keys, err := k.read()
	if err != nil {
		return err
	}
	defer f.Close()

	key := k.getKeyName(service, user)
	delete(keys, key)
	return k.write(f, keys)
}

func (k *File) getKeyName(service, user string) string {
	return fmt.Sprintf("%s:%s", service, user)
}

func (k *File) read() (*os.File, keys, error) {
	f, err := os.OpenFile(Keys.pathFile, os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, nil, err
	}

	keys := make(keys)
	if err := yaml.NewDecoder(f).Decode(&keys); err != nil {
		return nil, nil, err
	}
	return f, keys, nil
}

func (k *File) write(f *os.File, key keys) error {
	if err := f.Truncate(0); err != nil {
		return err
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	return yaml.NewEncoder(f).Encode(key)
}

func (k *File) encode(password string) (string, error) {
	block, err := aes.NewCipher(k.key)
	if err != nil {
		return "", err
	}

	plainText := []byte(password)
	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]

	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)
	return base64.RawStdEncoding.EncodeToString(cipherText), nil
}

func (k *File) decode(data string) (string, error) {
	cipherText, err := base64.RawStdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(k.key)
	if err != nil {
		return "", err
	}
	if len(cipherText) < aes.BlockSize {
		return "", errors.New("Ciphertext block size is too short")
	}
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)
	return string(cipherText), err
}

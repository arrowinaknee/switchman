package users

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const currentVersion = 1

var ErrPasswordEncoding = errors.New("incorrect password encoding")

func createPassword(password string) string {
	salt := randomHexString()
	hash := getPasswordHash(password, salt, currentVersion)
	return encodePassword(hash, salt, currentVersion)
}

func checkPassword(hashed string, password string) error {
	hash, salt, ver, err := decodePassword(hashed)
	if err != nil {
		return fmt.Errorf("cannot check password: %w", err)
	}

	cmp := hash
}

func getPasswordHash(password string, salt string, version int) [32]byte {
	concat := password + salt
	b := sha256.Sum256([]byte(concat))
	return b
}

func encodePassword(hash []byte, salt string, version int) string {
	hashStr := base64.RawStdEncoding.EncodeToString(hash)
	return fmt.Sprintf("v%x.%s.%s", version, hashStr, salt)
}

func decodePassword(encoded string) (hash [32]byte, salt string, version int, err error) {
	if encoded[0] != 'v' {
		err = fmt.Errorf("%w, missing version prefix", ErrPasswordEncoding)
		return
	}
	i := strings.IndexByte(encoded, '.')
	if i == -1 {
		err = fmt.Errorf("%w, missing separator after version", ErrPasswordEncoding)
		return
	}
	verStr := encoded[1:i]
	version, err = strconv.ParseInt(verStr, 10, 0)
	if err != nil {
		err = fmt.Errorf("%w, cannot parse version")
	}
}

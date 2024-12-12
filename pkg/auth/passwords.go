package auth

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
)

const (
	minVersion     = '1'
	currentVersion = '1'

	versionLen = 1
	hashLen    = 43
	saltLen    = 16
	encodedLen = versionLen + hashLen + saltLen
)

var ErrPasswordEncoding = errors.New("auth: password encoding")
var ErrPasswordMismatch = errors.New("auth: password mismatch")

// createEncodedPassword returns encoded password hash with salt
func createEncodedPassword(password string) string {
	salt := makeSalt()
	hash := getPasswordHash(password, salt, currentVersion)
	return encodePassword(hash, salt, currentVersion)
}

// checkPassword  and returns ErrPasswordMismatch if password doesn't match hashed
// or ErrPasswordEncoding if hashed password encoding is incorrect
func checkPassword(encoded string, password string) error {
	hash, salt, ver, err := decodePassword(encoded)
	if err != nil {
		return err
	}

	cmp := getPasswordHash(password, salt, ver)
	if !bytes.Equal(hash, cmp) {
		return ErrPasswordMismatch
	}
	return nil
}

// verifyEncodedPassword returns ErrPasswordEncoding if hashed password encoding is incorrect
func verifyEncodedPassword(hashed string) error {
	_, _, _, err := decodePassword(hashed)
	return err
}

func getPasswordHash(password string, salt string, version byte) []byte {
	_ = version // can be used later to change hashing algorithm
	concat := password + salt
	b := sha256.Sum256([]byte(concat))
	return b[:]
}

func encodePassword(hash []byte, salt string, version byte) string {
	hashStr := base64.RawStdEncoding.EncodeToString(hash)
	return fmt.Sprintf("%c%s%s", version, hashStr, salt)
}

func decodePassword(encoded string) (hash []byte, salt string, version byte, err error) {
	if len(encoded) != encodedLen {
		err = fmt.Errorf("%w: incorrect encoded length", ErrPasswordEncoding)
		return
	}
	var used int
	used, version, err = decodeVersion(encoded)
	if err != nil {
		return
	}
	encoded = encoded[used:]
	used, hash, err = decodeHash(encoded)
	if err != nil {
		return
	}
	encoded = encoded[used:]
	used, salt, err = decodeSalt(encoded)
	if err != nil {
		return
	}
	if used != len(encoded) {
		err = fmt.Errorf("%w: incorrect encoded length", ErrPasswordEncoding)
	}
	return
}

func decodeVersion(encoded string) (used int, version byte, err error) {
	version = encoded[0]
	if version < minVersion || version > currentVersion {
		return 0, 0, fmt.Errorf("%w: version too high", ErrPasswordEncoding)
	}
	return versionLen, version, nil
}

func decodeHash(encoded string) (used int, hash []byte, err error) {
	hash, err = base64.RawStdEncoding.DecodeString(encoded[:hashLen])
	if err != nil {
		return 0, nil, fmt.Errorf("%w: cannot decode hash: %w", ErrPasswordEncoding, err)
	}
	return hashLen, hash, nil
}

func decodeSalt(encoded string) (used int, salt string, err error) {
	salt = encoded[:saltLen]
	return saltLen, salt, nil
}

func makeSalt() string {
	buf := make([]byte, saltLen/2)
	random.Read(buf) // random.Read() never returns errors

	return hex.EncodeToString(buf)
}

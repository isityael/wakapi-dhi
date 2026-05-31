package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/duke-git/lancet/v2/mathutil"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

var md5Regex = regexp.MustCompile(`^[a-f0-9]{32}$`)

const (
	argon2IDMemory     = 64 * 1024
	argon2IDIterations = 1
	argon2IDSaltLength = 16
	argon2IDKeyLength  = 32
)

type argon2IDParams struct {
	Memory      uint32
	Time        uint32
	Parallelism uint8
	KeyLength   uint32
}

func ExtractBasicAuth(r *http.Request) (username, password string, err error) {
	authHeader := strings.Split(r.Header.Get("Authorization"), " ")
	if len(authHeader) != 2 || authHeader[0] != "Basic" {
		return username, password, errors.New("failed to extract API key")
	}

	hash, err := base64.StdEncoding.DecodeString(authHeader[1])
	userKey := strings.TrimSpace(string(hash))
	if err != nil {
		return username, password, err
	}

	re := regexp.MustCompile(`^(.+):(.+)$`)
	groups := re.FindAllStringSubmatch(userKey, -1)
	if len(groups) == 0 || len(groups[0]) != 3 {
		return username, password, errors.New("failed to parse user agent string")
	}
	username, password = groups[0][1], groups[0][2]
	return username, password, err
}

func ExtractBearerAuth(r *http.Request) (key string, err error) {
	authHeader := strings.Split(r.Header.Get("Authorization"), " ")
	if len(authHeader) != 2 || (authHeader[0] != "Basic" && authHeader[0] != "Bearer") {
		return key, errors.New("failed to extract API key")
	}

	keyBytes, err := base64.StdEncoding.DecodeString(authHeader[1])
	return string(keyBytes), err
}

// password hashing

func ComparePassword(hashed, plain, pepper string) bool {
	if hashed[0:10] == "$argon2id$" {
		return CompareArgon2Id(hashed, plain, pepper)
	}
	return CompareBcrypt(hashed, plain, pepper)
}

func HashPassword(plain, pepper string) (string, error) {
	return HashArgon2Id(plain, pepper)
}

func CompareBcrypt(hashed, plain, pepper string) bool {
	plainPepperedPassword := []byte(strings.TrimSpace(plain) + pepper)
	err := bcrypt.CompareHashAndPassword([]byte(hashed), plainPepperedPassword)
	return err == nil
}

func HashBcrypt(plain, pepper string) (string, error) {
	plainPepperedPassword := []byte(strings.TrimSpace(plain) + pepper)
	bytes, err := bcrypt.GenerateFromPassword(plainPepperedPassword, bcrypt.DefaultCost)
	if err == nil {
		return string(bytes), nil
	}
	return "", err
}

func CompareArgon2Id(hashed, plain, pepper string) bool {
	plainPepperedPassword := strings.TrimSpace(plain) + pepper
	params, salt, key, err := decodeArgon2IDHash(hashed)
	if err != nil {
		return false
	}

	otherKey := argon2.IDKey([]byte(plainPepperedPassword), salt, params.Time, params.Memory, params.Parallelism, params.KeyLength)
	if subtle.ConstantTimeEq(int32(len(key)), int32(len(otherKey))) == 0 {
		return false
	}
	return subtle.ConstantTimeCompare(key, otherKey) == 1
}

func HashArgon2Id(plain, pepper string) (string, error) {
	plainPepperedPassword := strings.TrimSpace(plain) + pepper
	params := argon2IDParams{
		Memory:      argon2IDMemory,
		Time:        argon2IDIterations,
		Parallelism: uint8(mathutil.Min[int](runtime.NumCPU(), 255)),
		KeyLength:   argon2IDKeyLength,
	}

	if params.Parallelism == 0 { // https://github.com/muety/wakapi/issues/866
		params.Parallelism = 1
	}

	salt := make([]byte, argon2IDSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	key := argon2.IDKey([]byte(plainPepperedPassword), salt, params.Time, params.Memory, params.Parallelism, params.KeyLength)
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		params.Memory,
		params.Time,
		params.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

func decodeArgon2IDHash(hash string) (params *argon2IDParams, salt, key []byte, err error) {
	vals := strings.Split(hash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, errors.New("invalid argon2id hash")
	}
	if vals[1] != "argon2id" {
		return nil, nil, nil, errors.New("incompatible argon2 variant")
	}

	version, err := parsePrefixedUint32(vals[2], "v=")
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, errors.New("incompatible argon2 version")
	}

	settings := strings.Split(vals[3], ",")
	if len(settings) != 3 {
		return nil, nil, nil, errors.New("invalid argon2id parameters")
	}

	memory, err := parsePrefixedUint32(settings[0], "m=")
	if err != nil {
		return nil, nil, nil, err
	}
	time, err := parsePrefixedUint32(settings[1], "t=")
	if err != nil {
		return nil, nil, nil, err
	}
	parallelism, err := parsePrefixedUint32(settings[2], "p=")
	if err != nil {
		return nil, nil, nil, err
	}
	if parallelism > 255 {
		return nil, nil, nil, errors.New("argon2id parallelism exceeds uint8")
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	key, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}

	return &argon2IDParams{
		Memory:      memory,
		Time:        time,
		Parallelism: uint8(parallelism),
		KeyLength:   uint32(len(key)),
	}, salt, key, nil
}

func parsePrefixedUint32(value, prefix string) (uint32, error) {
	if !strings.HasPrefix(value, prefix) {
		return 0, fmt.Errorf("missing %q prefix", prefix)
	}
	parsed, err := strconv.ParseUint(strings.TrimPrefix(value, prefix), 10, 32)
	return uint32(parsed), err
}

package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeArgon2IDHash(t *testing.T) {
	params, salt, key, err := decodeArgon2IDHash("$argon2id$v=19$m=65536,t=3,p=4$c29tZXNhbHQ$YWJjZGVmZw")

	require.NoError(t, err)
	assert.EqualValues(t, 65536, params.Memory)
	assert.EqualValues(t, 3, params.Time)
	assert.EqualValues(t, 4, params.Parallelism)
	assert.Equal(t, []byte("somesalt"), salt)
	assert.Equal(t, []byte("abcdefg"), key)
}

func TestHashArgon2IdVerifiesAndUsesPHCFormat(t *testing.T) {
	hash, err := HashArgon2Id("correct horse battery staple", "pepper")

	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(hash, "$argon2id$v=19$m=65536,t=1,p="))
	assert.True(t, CompareArgon2Id(hash, "correct horse battery staple", "pepper"))
	assert.False(t, CompareArgon2Id(hash, "wrong", "pepper"))
	assert.False(t, CompareArgon2Id(hash, "correct horse battery staple", "wrong-pepper"))
}

func TestCompareArgon2IdRejectsMalformedHashes(t *testing.T) {
	assert.False(t, CompareArgon2Id("$argon2id$v=19$m=65536,t=3,p=4$c29tZXNhbHQ", "plain", "pepper"))
	assert.False(t, CompareArgon2Id("$argon2i$v=19$m=65536,t=3,p=4$c29tZXNhbHQ$YWJjZGVmZw", "plain", "pepper"))
}

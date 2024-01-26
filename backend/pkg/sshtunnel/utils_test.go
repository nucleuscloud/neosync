package sshtunnel

import (
	"testing"

	"github.com/zeebo/assert"
)

const (
	encryptedPrivateKey = `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAACmFlczI1Ni1jdHIAAAAGYmNyeXB0AAAAGAAAABDcxXuNyz
EyQ3fS7uiTcfvDAAAAGAAAAAEAAAAzAAAAC3NzaC1lZDI1NTE5AAAAIHRde4TANOm21rV4
hyHkZjPHFJazaxZHd9M/TzchhVKhAAAAoGQ2S553lBIdQHDHwsC+ySbc8PShkW2tmF9X2h
LHW/Zvmd4uy2/jg7kWMnWPfkUkIytjD0Llo+o6yTq3wfaGfOkn8M57NcwGvdvHoCIswbv/
COyG2jmUCxomIKi0qDxzDnp22ELGKpdEDTjit1d8aNwjWZU73AfyPwulhTa9H/uxao1Qat
LqqnUvkQBvhk/q8M2CpbmDwBXJ8x3IVXOx/dQ=
-----END OPENSSH PRIVATE KEY-----`
	encryptedPrivateKeyPass = "foobar"

	unencryptedPrivateKey = `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACCVXq9QVRO6CLsUemebj/8gcFJkw4x6dmQXlzrZ0J4opgAAAJhALCAYQCwg
GAAAAAtzc2gtZWQyNTUxOQAAACCVXq9QVRO6CLsUemebj/8gcFJkw4x6dmQXlzrZ0J4opg
AAAEAVL3RnsSDw63JV+ATzXYtmfIW6EMY4PQ2227MsSYEUdpVer1BVE7oIuxR6Z5uP/yBw
UmTDjHp2ZBeXOtnQniimAAAAEHRlc3RAZXhhbXBsZS5jb20BAgMEBQ==
-----END OPENSSH PRIVATE KEY-----`
)

func Test_GetPrivateKeyAuthMethod(t *testing.T) {
	out, err := GetPrivateKeyAuthMethod([]byte(encryptedPrivateKey), ptr(encryptedPrivateKeyPass))
	assert.NoError(t, err)
	assert.NotNil(t, out)

	out, err = GetPrivateKeyAuthMethod([]byte(encryptedPrivateKey), ptr("badpassword"))
	assert.Error(t, err)
	assert.Nil(t, out)

	out, err = GetPrivateKeyAuthMethod([]byte("bad key"), ptr(encryptedPrivateKeyPass))
	assert.Error(t, err)
	assert.Nil(t, out)

	out, err = GetPrivateKeyAuthMethod([]byte(unencryptedPrivateKey), nil)
	assert.NoError(t, err)
	assert.NotNil(t, out)

	out, err = GetPrivateKeyAuthMethod([]byte("bad key"), nil)
	assert.Error(t, err)
	assert.Nil(t, out)
}

func ptr[T any](val T) *T {
	return &val
}

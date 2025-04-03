package sshtunnel

import (
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
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
	unencryptedPublicKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJVer1BVE7oIuxR6Z5uP/yBwUmTDjHp2ZBeXOtnQniim test@example.com"
)

func Test_getPrivateKeyAuthMethod(t *testing.T) {
	out, err := getPrivateKeyAuthMethod([]byte(encryptedPrivateKey), ptr(encryptedPrivateKeyPass))
	assert.NoError(t, err)
	assert.NotNil(t, out)

	out, err = getPrivateKeyAuthMethod([]byte(encryptedPrivateKey), ptr("badpassword"))
	assert.Error(t, err)
	assert.Nil(t, out)

	out, err = getPrivateKeyAuthMethod([]byte("bad key"), ptr(encryptedPrivateKeyPass))
	assert.Error(t, err)
	assert.Nil(t, out)

	out, err = getPrivateKeyAuthMethod([]byte(unencryptedPrivateKey), nil)
	assert.NoError(t, err)
	assert.NotNil(t, out)

	out, err = getPrivateKeyAuthMethod([]byte("bad key"), nil)
	assert.Error(t, err)
	assert.Nil(t, out)
}

func ptr[T any](val T) *T {
	return &val
}

func Test_parseSshKey(t *testing.T) {
	pk, err := parseSshKey(unencryptedPublicKey)
	assert.NoError(t, err)
	assert.NotNil(t, pk)

	pk, err = parseSshKey("bad key")
	assert.Error(t, err)
	assert.Nil(t, pk)
}

func Test_getTunnelAuthMethodFromSshConfig(t *testing.T) {
	out, err := getTunnelAuthMethodFromSshConfig(nil)
	assert.NoError(t, err)
	assert.Nil(t, out)

	out, err = getTunnelAuthMethodFromSshConfig(&mgmtv1alpha1.SSHAuthentication{})
	assert.NoError(t, err)
	assert.Nil(t, out)

	out, err = getTunnelAuthMethodFromSshConfig(&mgmtv1alpha1.SSHAuthentication{
		AuthConfig: &mgmtv1alpha1.SSHAuthentication_Passphrase{
			Passphrase: &mgmtv1alpha1.SSHPassphrase{
				Value: "foo",
			},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, out)

	out, err = getTunnelAuthMethodFromSshConfig(&mgmtv1alpha1.SSHAuthentication{
		AuthConfig: &mgmtv1alpha1.SSHAuthentication_PrivateKey{
			PrivateKey: &mgmtv1alpha1.SSHPrivateKey{
				Value:      encryptedPrivateKey,
				Passphrase: ptr(encryptedPrivateKeyPass),
			},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, out)

	out, err = getTunnelAuthMethodFromSshConfig(&mgmtv1alpha1.SSHAuthentication{
		AuthConfig: &mgmtv1alpha1.SSHAuthentication_PrivateKey{
			PrivateKey: &mgmtv1alpha1.SSHPrivateKey{
				Value:      encryptedPrivateKey,
				Passphrase: ptr("badpass"),
			},
		},
	})
	assert.Error(t, err)
	assert.Nil(t, out)
}

func Test_getSshAddr(t *testing.T) {
	t.Run("with port", func(t *testing.T) {
		actual := getSshAddr(&mgmtv1alpha1.SSHTunnel{
			Host: "localhost",
			Port: 2222,
		})
		assert.Equal(t, "localhost:2222", actual)
	})
	t.Run("without port", func(t *testing.T) {
		actual := getSshAddr(&mgmtv1alpha1.SSHTunnel{
			Host: "localhost",
		})
		assert.Equal(t, "localhost", actual)
	})
}

package controllers

import (
	"testing"

	neosyncdevv1alpha1 "github.com/nucleuscloud/neosync/k8s-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestDoesSqlConnUseSecret(t *testing.T) {
	assert.False(t, doesSqlConnUseSecret(nil, "foo"))
	assert.False(t, doesSqlConnUseSecret(&neosyncdevv1alpha1.SqlConnection{}, "foo"))
	assert.False(t, doesSqlConnUseSecret(&neosyncdevv1alpha1.SqlConnection{
		Spec: neosyncdevv1alpha1.SqlConnectionSpec{},
	}, "foo"))
	assert.False(t, doesSqlConnUseSecret(&neosyncdevv1alpha1.SqlConnection{
		Spec: neosyncdevv1alpha1.SqlConnectionSpec{
			Url: neosyncdevv1alpha1.SqlConnectionUrl{},
		},
	}, "foo"))
	assert.False(t, doesSqlConnUseSecret(&neosyncdevv1alpha1.SqlConnection{
		Spec: neosyncdevv1alpha1.SqlConnectionSpec{
			Url: neosyncdevv1alpha1.SqlConnectionUrl{
				ValueFrom: &neosyncdevv1alpha1.SqlConnectionUrlSource{},
			},
		},
	}, "foo"))
	assert.False(t, doesSqlConnUseSecret(&neosyncdevv1alpha1.SqlConnection{
		Spec: neosyncdevv1alpha1.SqlConnectionSpec{
			Url: neosyncdevv1alpha1.SqlConnectionUrl{
				ValueFrom: &neosyncdevv1alpha1.SqlConnectionUrlSource{
					SecretKeyRef: &neosyncdevv1alpha1.ConfigSelector{},
				},
			},
		},
	}, "foo"))
	assert.False(t, doesSqlConnUseSecret(&neosyncdevv1alpha1.SqlConnection{
		Spec: neosyncdevv1alpha1.SqlConnectionSpec{
			Url: neosyncdevv1alpha1.SqlConnectionUrl{
				ValueFrom: &neosyncdevv1alpha1.SqlConnectionUrlSource{
					SecretKeyRef: &neosyncdevv1alpha1.ConfigSelector{
						Name: "foo2",
					},
				},
			},
		},
	}, "foo"))
	assert.True(t, doesSqlConnUseSecret(&neosyncdevv1alpha1.SqlConnection{
		Spec: neosyncdevv1alpha1.SqlConnectionSpec{
			Url: neosyncdevv1alpha1.SqlConnectionUrl{
				ValueFrom: &neosyncdevv1alpha1.SqlConnectionUrlSource{
					SecretKeyRef: &neosyncdevv1alpha1.ConfigSelector{
						Name: "foo",
					},
				},
			},
		},
	}, "foo"))
}

func TestGenerateSha256Hash(t *testing.T) {
	input := "foo"
	assert.Equal(
		t,
		generateSha256Hash([]byte(input)),
		"2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
	)
}

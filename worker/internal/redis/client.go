package neosync_redis

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"github.com/redis/go-redis/v9"
)

func GetRedisClient(redisConfig *shared.RedisConfig) (redis.UniversalClient, error) {
	if redisConfig == nil {
		return nil, nil
	}

	var tlsConf *tls.Config
	if redisConfig.Tls != nil && redisConfig.Tls.Enabled {
		tlsc, err := getTlsConfig(redisConfig.Tls)
		if err != nil {
			return nil, err
		}
		tlsConf = tlsc
	}

	var redisDB int
	var user string
	var pass string
	var addrs []string

	redisUrl, err := url.Parse(redisConfig.Url)
	if err != nil {
		return nil, err
	}

	if redisUrl.Scheme == "tcp" {
		redisUrl.Scheme = "redis"
	}

	rurl, err := redis.ParseURL(redisUrl.String())
	if err != nil {
		return nil, err
	}

	addrs = append(addrs, rurl.Addr)
	redisDB = rurl.DB
	user = rurl.Username
	pass = rurl.Password

	var client redis.UniversalClient
	opts := &redis.UniversalOptions{
		Addrs:     addrs,
		DB:        redisDB,
		Username:  user,
		Password:  pass,
		TLSConfig: tlsConf,
	}

	switch redisConfig.Kind {
	case "simple":
		client = redis.NewClient(opts.Simple())
	case "cluster":
		client = redis.NewClusterClient(opts.Cluster())
	case "failover":
		opts.MasterName = *redisConfig.Master
		client = redis.NewFailoverClient(opts.Failover())
	default:
		return nil, fmt.Errorf("invalid redis kind: %s", redisConfig.Kind)
	}

	return client, nil
}

func defaultTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
}

func getTlsConfig(c *shared.RedisTlsConfig) (*tls.Config, error) {
	var tlsConf *tls.Config
	initConf := func() {
		if tlsConf != nil {
			return
		}
		tlsConf = defaultTLSConfig()
	}

	if *c.RootCertAuthority != "" && *c.RootCertAuthorityFile != "" {
		return nil, errors.New("only one field between root_cas and root_cas_file can be specified")
	}

	if c.RootCertAuthorityFile != nil && *c.RootCertAuthorityFile != "" {
		caCert, err := readFile(*c.RootCertAuthorityFile)
		if err != nil {
			return nil, err
		}
		initConf()
		tlsConf.RootCAs = x509.NewCertPool()
		tlsConf.RootCAs.AppendCertsFromPEM(caCert)
	}

	if c.RootCertAuthority != nil && *c.RootCertAuthority != "" {
		initConf()
		tlsConf.RootCAs = x509.NewCertPool()
		tlsConf.RootCAs.AppendCertsFromPEM([]byte(*c.RootCertAuthority))
	}

	if c.EnableRenegotiation {
		initConf()
		tlsConf.Renegotiation = tls.RenegotiateFreelyAsClient
	}

	if c.SkipCertVerify {
		initConf()
		tlsConf.InsecureSkipVerify = true
	}

	if c.Enabled && tlsConf == nil {
		initConf()
	}

	return tlsConf, nil
}

func readFile(filename string) ([]byte, error) {
	// Open the file with RDONLY flag
	file, err := os.OpenFile(filename, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read all bytes from the file
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return data, nil
}

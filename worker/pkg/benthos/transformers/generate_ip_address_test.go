package transformers

import (
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

func TestPublicIPv4Generation(t *testing.T) {
	randomizer := rng.New(1234)
	ip, err := generateIpAddress(randomizer, IpV4_Public, 100)
	require.NoError(t, err)
	require.NotEmpty(t, ip)

	parsedIP := net.ParseIP(ip)
	require.NotNil(t, parsedIP, "should be valid IP")
	assert.False(t, isReservedIP(parsedIP), "should not be reserved IP")
}

func TestPrivateAIPv4Generation(t *testing.T) {
	randomizer := rng.New(1234)
	ip, err := generateIpAddress(randomizer, IpV4_PrivateA, 100)
	require.NoError(t, err)
	require.NotEmpty(t, ip)

	assert.True(t, strings.HasPrefix(ip, "10."))
	assertIPInRange(t, ip, "10.0.0.0", "10.255.255.255")
}

func TestPrivateBIPv4Generation(t *testing.T) {
	randomizer := rng.New(1234)
	ip, err := generateIpAddress(randomizer, IpV4_PrivateB, 100)
	require.NoError(t, err)
	require.NotEmpty(t, ip)

	assert.True(t, strings.HasPrefix(ip, "172."))
	parts := strings.Split(ip, ".")
	secondOctet := parseInt(t, parts[1])
	assert.GreaterOrEqual(t, secondOctet, 16)
	assert.LessOrEqual(t, secondOctet, 31)
	assertIPInRange(t, ip, "172.16.0.0", "172.31.255.255")
}

func TestPrivateCIPv4Generation(t *testing.T) {
	randomizer := rng.New(1234)
	ip, err := generateIpAddress(randomizer, IpV4_PrivateC, 100)
	require.NoError(t, err)
	require.NotEmpty(t, ip)

	assert.True(t, strings.HasPrefix(ip, "192.168."))
	assertIPInRange(t, ip, "192.168.0.0", "192.168.255.255")
}

func TestLinkLocalIPv4Generation(t *testing.T) {
	randomizer := rng.New(1234)
	ip, err := generateIpAddress(randomizer, IpV4_LinkLocal, 100)
	require.NoError(t, err)
	require.NotEmpty(t, ip)

	assert.True(t, strings.HasPrefix(ip, "169.254."))
	assertIPInRange(t, ip, "169.254.0.0", "169.254.255.255")
}

func TestMulticastIPv4Generation(t *testing.T) {
	randomizer := rng.New(1234)
	ip, err := generateIpAddress(randomizer, IpV4_Multicast, 100)
	require.NoError(t, err)
	require.NotEmpty(t, ip)

	firstOctet := parseInt(t, strings.Split(ip, ".")[0])
	assert.GreaterOrEqual(t, firstOctet, 224)
	assert.LessOrEqual(t, firstOctet, 239)
	assertIPInRange(t, ip, "224.0.0.0", "239.255.255.255")
}

func TestLoopbackIPv4Generation(t *testing.T) {
	randomizer := rng.New(1234)
	ip, err := generateIpAddress(randomizer, IpV4_Loopback, 100)
	require.NoError(t, err)
	require.NotEmpty(t, ip)

	assert.True(t, strings.HasPrefix(ip, "127."))
	assertIPInRange(t, ip, "127.0.0.0", "127.255.255.255")
}

func TestIPv6Generation(t *testing.T) {
	randomizer := rng.New(1234)
	ip, err := generateIpAddress(randomizer, IpV4_V6, 100)
	require.NoError(t, err)
	require.NotEmpty(t, ip)

	parsedIP := net.ParseIP(ip)
	require.NotNil(t, parsedIP, "should be valid IPv6")

	groups := strings.Split(ip, ":")
	assert.Equal(t, 8, len(groups), "should have 8 groups")
	for _, group := range groups {
		assert.Len(t, group, 4, "each group should be 4 characters")
		// Verify each group is valid hexadecimal
		_, err := parseHexGroup(group)
		assert.NoError(t, err, "each group should be valid hexadecimal")
	}
}

func TestInvalidIPTypeError(t *testing.T) {
	randomizer := rng.New(1234)
	_, err := generateIpAddress(randomizer, IpType("invalid"), 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported IPv4 type")
}

func TestIPConversion(t *testing.T) {
	testCases := map[string]string{
		"public ip": "8.8.8.8",
		"private A": "10.0.0.1",
		"private B": "172.16.0.1",
		"private C": "192.168.1.1",
		"min value": "0.0.0.0",
		"max value": "255.255.255.255",
	}

	for name, ipStr := range testCases {
		t.Run(name, func(t *testing.T) {
			originalIP := net.ParseIP(ipStr)
			require.NotNil(t, originalIP)

			intIP := ipToInt64(originalIP)
			convertedIP := int64ToIP(intIP)

			assert.Equal(t, originalIP.To4().String(), convertedIP.String())
		})
	}
}

func TestIPRangeGeneration(t *testing.T) {
	randomizer := rng.New(1234)

	t.Run("private A range", func(t *testing.T) {
		start := net.ParseIP("10.0.0.0")
		end := net.ParseIP("10.255.255.255")
		ip := generateIPInRange(randomizer, start, end)
		assert.True(t, strings.HasPrefix(ip, "10."))
		assertIPInRange(t, ip, "10.0.0.0", "10.255.255.255")
	})

	t.Run("private B range", func(t *testing.T) {
		start := net.ParseIP("172.16.0.0")
		end := net.ParseIP("172.31.255.255")
		ip := generateIPInRange(randomizer, start, end)
		assert.True(t, strings.HasPrefix(ip, "172."))
		parts := strings.Split(ip, ".")
		secondOctet := parseInt(t, parts[1])
		assert.GreaterOrEqual(t, secondOctet, 16)
		assert.LessOrEqual(t, secondOctet, 31)
	})

	t.Run("single IP range", func(t *testing.T) {
		start := net.ParseIP("192.168.1.1")
		end := net.ParseIP("192.168.1.1")
		ip := generateIPInRange(randomizer, start, end)
		assert.Equal(t, "192.168.1.1", ip)
	})
}

func Test_IpV4Public_NoOptions(t *testing.T) {
	mapping := `root = generate_ip()`
	ip, err := bloblang.Parse(mapping)
	require.NoError(t, err)
	require.NotEmpty(t, ip)

	res, err := ip.Query(nil)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	resStr, ok := res.(string)
	require.True(t, ok)
	require.NotEmpty(t, resStr)

	parsedIP := net.ParseIP(resStr)
	require.NotNil(t, parsedIP, "should be valid IP")
	assert.False(t, isReservedIP(parsedIP), "should not be reserved IP")
}

func Test_IpV4PrivateA(t *testing.T) {
	ipType := string(IpV4_PrivateA)
	mapping := fmt.Sprintf(`root = generate_ip(ip_type:%q)`, ipType)
	ip, err := bloblang.Parse(mapping)
	require.NoError(t, err)
	require.NotEmpty(t, ip)

	res, err := ip.Query(nil)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	resStr, ok := res.(string)
	require.True(t, ok)
	require.NotEmpty(t, resStr)

	parsedIP := net.ParseIP(resStr)
	require.NotNil(t, parsedIP, "should be valid IP")
	assert.True(t, isReservedIP(parsedIP), "should be reserved IP since it's private")
}

func Test_IpV6_Generation(t *testing.T) {
	ipType := string(IpV4_V6)
	mapping := fmt.Sprintf(`root = generate_ip(ip_type:%q)`, ipType)
	ip, err := bloblang.Parse(mapping)
	require.NoError(t, err)
	require.NotEmpty(t, ip)

	res, err := ip.Query(nil)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	resStr, ok := res.(string)
	require.True(t, ok)
	require.NotEmpty(t, resStr)

	parsedIP := net.ParseIP(resStr)
	require.NotNil(t, parsedIP, "should be valid IP")

	groups := strings.Split(resStr, ":")
	assert.Equal(t, 8, len(groups), "should have 8 groups")
	for _, group := range groups {
		assert.Len(t, group, 4, "each group should be 4 characters")
		_, err := parseHexGroup(group)
		assert.NoError(t, err, "each group should be valid hexadecimal")
	}
}

func Test_IpV4PrivateB(t *testing.T) {
	ipType := string(IpV4_PrivateB)
	mapping := fmt.Sprintf(`root = generate_ip(ip_type:%q)`, ipType)
	ip, err := bloblang.Parse(mapping)
	require.NoError(t, err)
	require.NotEmpty(t, ip)

	res, err := ip.Query(nil)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	resStr, ok := res.(string)
	require.True(t, ok)
	require.NotEmpty(t, resStr)

	parsedIP := net.ParseIP(resStr)
	require.NotNil(t, parsedIP, "should be valid IP")
	assert.True(t, isReservedIP(parsedIP), "should be reserved IP since it's private")
}

func assertIPInRange(t *testing.T, ip, start, end string) {
	parsedIP := net.ParseIP(ip)
	startIP := net.ParseIP(start)
	endIP := net.ParseIP(end)
	require.NotNil(t, parsedIP)
	require.NotNil(t, startIP)
	require.NotNil(t, endIP)

	assert.True(t, inRange(parsedIP, startIP, endIP),
		"IP %s should be in range %s-%s", ip, start, end)
}

func parseInt(t *testing.T, s string) int {
	var n int
	for _, ch := range s {
		n = n*10 + int(ch-'0')
	}
	return n
}

func parseHexGroup(s string) (int64, error) {
	return parseInt64(s, 16)
}

func parseInt64(s string, base int) (int64, error) {
	var n int64
	for _, ch := range s {
		n = n * int64(base)
		switch {
		case ch >= '0' && ch <= '9':
			n += int64(ch - '0')
		case ch >= 'a' && ch <= 'f':
			n += int64(ch - 'a' + 10)
		case ch >= 'A' && ch <= 'F':
			n += int64(ch - 'A' + 10)
		default:
			return 0, fmt.Errorf("invalid character: %c", ch)
		}
	}
	return n, nil
}

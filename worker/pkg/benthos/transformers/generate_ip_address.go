package transformers

import (
	"bytes"
	"fmt"
	"net"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateIpAddress

type IpType string

const (
	IpV4_Public    IpType = "GENERATE_IP_ADDRESS_TYPE_V4_PUBLIC"
	IpV4_PrivateA  IpType = "GENERATE_IP_ADDRESS_TYPE_V4_PRIVATE_A"
	IpV4_PrivateB  IpType = "GENERATE_IP_ADDRESS_TYPE_V4_PRIVATE_B"
	IpV4_PrivateC  IpType = "GENERATE_IP_ADDRESS_TYPE_V4_PRIVATE_C"
	IpV4_LinkLocal IpType = "GENERATE_IP_ADDRESS_TYPE_V4_LINK_LOCAL"
	IpV4_Multicast IpType = "GENERATE_IP_ADDRESS_TYPE_V4_MULTICAST"
	IpV4_Loopback  IpType = "GENERATE_IP_ADDRESS_TYPE_V4_LOOPBACK"
	IpV4_V6        IpType = "GENERATE_IP_ADDRESS_TYPE_V6"
)

// Defined here -> https://www.meridianoutpost.com/resources/articles/IP-classes.php
// And here -> https://www.techtarget.com/whatis/definition/private-IP-addresshttps://www.techtarget.com/whatis/definition/private-IP-address
var ipv4Ranges = map[IpType]struct {
	start net.IP
	end   net.IP
}{
	IpV4_PrivateA:  {net.ParseIP("10.0.0.0"), net.ParseIP("10.255.255.255")},
	IpV4_PrivateB:  {net.ParseIP("172.16.0.0"), net.ParseIP("172.31.255.255")},
	IpV4_PrivateC:  {net.ParseIP("192.168.0.0"), net.ParseIP("192.168.255.255")},
	IpV4_LinkLocal: {net.ParseIP("169.254.0.0"), net.ParseIP("169.254.255.255")},
	IpV4_Multicast: {net.ParseIP("224.0.0.0"), net.ParseIP("239.255.255.255")},
	IpV4_Loopback:  {net.ParseIP("127.0.0.0"), net.ParseIP("127.255.255.255")},
}

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Generates IPv4 or IPv6 addresses with support for different network classes.").
		Category("string").
		Param(bloblang.NewInt64Param("max_length").Default(100000).Description("Specifies the maximum length for the generated data. This field ensures that the output does not exceed a certain number of characters.")).
		Param(bloblang.NewStringParam("ip_type").Default(string(IpV4_Public)).Description("IP type to generate.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("Optional seed for deterministic generation"))

	err := bloblang.RegisterFunctionV2(
		"generate_ip",
		spec,
		func(args *bloblang.ParsedParams) (bloblang.Function, error) {
			maxLength, err := args.GetInt64("max_length")
			if err != nil {
				return nil, err
			}
			ipType, err := args.GetString("ip_type")
			if err != nil {
				return nil, err
			}

			seedArg, err := args.GetOptionalInt64("seed")
			if err != nil {
				return nil, err
			}

			seed, err := transformer_utils.GetSeedOrDefault(seedArg)
			if err != nil {
				return nil, err
			}

			randomizer := rng.New(seed)

			return func() (any, error) {
				return generateIpAddress(randomizer, IpType(ipType), maxLength)
			}, nil
		},
	)

	if err != nil {
		panic(err)
	}
}

func NewGenerateIpAddressOptsFromConfig(
	config *mgmtv1alpha1.GenerateIpAddress,
	maxlength *int64,
) (*GenerateIpAddressOpts, error) {
	if config == nil {
		return NewGenerateIpAddressOpts(nil, nil, nil)
	}

	var ipType *string

	defaultIpType := string(IpV4_Public)
	ipType = &defaultIpType
	if config.IpType != nil {
		v := config.IpType.String()
		ipType = &v
	}

	return NewGenerateIpAddressOpts(maxlength, ipType, nil)
}

func (t *GenerateIpAddress) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateIpAddressOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}
	return generateIpAddress(parsedOpts.randomizer, IpType(parsedOpts.ipType), parsedOpts.maxLength)
}

func generateIpAddress(randomizer rng.Rand, ipType IpType, maxLength int64) (string, error) {
	var ip string
	var err error

	if ipType == IpV4_V6 {
		ip = generateIPv6Address(randomizer)
	} else {
		ip, err = generateIPv4Address(randomizer, ipType)
	}

	if err != nil {
		return "", err
	}

	if maxLength <= 0 {
		return ip, nil
	}

	if int64(len(ip)) > maxLength {
		return ip[:maxLength], nil
	}

	return ip, nil
}

func generateIPv4Address(randomizer rng.Rand, ipType IpType) (string, error) {
	if ipType == IpV4_Public {
		return generatePublicIPv4(randomizer)
	}

	ipRange, exists := ipv4Ranges[ipType]
	if !exists {
		return "", fmt.Errorf("unsupported IPv4 type: %s", ipType)
	}

	return generateIPInRange(randomizer, ipRange.start, ipRange.end), nil
}

func generateIPv6Address(randomizer rng.Rand) string {
	groups := make([]string, 8)
	for i := 0; i < 8; i++ {
		groups[i] = fmt.Sprintf("%04x", randomizer.Intn(65536))
	}
	return strings.Join(groups, ":")
}

func generatePublicIPv4(randomizer rng.Rand) (string, error) {
	maxAttempts := 1000
	for i := 0; i < maxAttempts; i++ {
		ip := make(net.IP, 4)
		for i := range ip {
			ip[i] = byte(randomizer.Intn(256))
		}

		if !isReservedIP(ip) {
			return ip.String(), nil
		}
	}
	return "", fmt.Errorf("failed to generate public IP after %d attempts", maxAttempts)
}
func isReservedIP(ip net.IP) bool {
	for _, ipRange := range ipv4Ranges {
		if inRange(ip, ipRange.start, ipRange.end) {
			return true
		}
	}
	// Check experimental/reserved range (240.0.0.0/4)
	experimentalStart := net.ParseIP("240.0.0.0")
	experimentalEnd := net.ParseIP("255.255.255.255")
	return inRange(ip, experimentalStart, experimentalEnd)
}

func generateIPInRange(randomizer rng.Rand, start, end net.IP) string {
	startInt := ipToInt64(start)
	endInt := ipToInt64(end)
	rangeSize := endInt - startInt + 1
	random := startInt + randomizer.Int63n(rangeSize)
	return int64ToIP(random).String()
}

// normalizes IPs into the same format
func normalizeIP(ip net.IP) net.IP {
	if ipv4 := ip.To4(); ipv4 != nil {
		// return ipv4 as 4 bytes
		return ipv4
	}
	// format for ipv6
	return ip.To16()
}

func inRange(ip, start, end net.IP) bool {
	// Normalize all IPs to ensure consistent comparison
	ip = normalizeIP(ip)
	start = normalizeIP(start)
	end = normalizeIP(end)

	// Ensure all IPs are in the same format (either all IPv4 or all IPv6)
	if ip == nil || start == nil || end == nil {
		return false
	}

	// Ensure we're comparing IPs of the same version
	if len(ip) != len(start) || len(ip) != len(end) {
		return false
	}

	return bytes.Compare(ip, start) >= 0 && bytes.Compare(ip, end) <= 0
}

func ipToInt64(ip net.IP) int64 {
	ip = ip.To4()
	return int64(ip[0])<<24 | int64(ip[1])<<16 | int64(ip[2])<<8 | int64(ip[3])
}

func int64ToIP(n int64) net.IP {
	ip := make(net.IP, 4)
	ip[0] = byte(n >> 24)
	ip[1] = byte(n >> 16)
	ip[2] = byte(n >> 8)
	ip[3] = byte(n)
	return ip
}

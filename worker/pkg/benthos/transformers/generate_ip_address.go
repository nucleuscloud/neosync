package transformers

import (
	"bytes"
	"fmt"
	"net"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateIpAddress

type IpVersion string

const (
	IpVersion_V4 IpVersion = "GENERATE_IP_ADDRESS_VERSION_V4"
	IpVersion_V6 IpVersion = "GENERATE_IP_ADDRESS_VERSION_V6"
)

type IpV4Class string

const (
	IpV4Class_Public    IpV4Class = "GENERATE_IP_ADDRESS_CLASS_PUBLIC"
	IpV4Class_PrivateA  IpV4Class = "GENERATE_IP_ADDRESS_CLASS_PRIVATE_A"
	IpV4Class_PrivateB  IpV4Class = "GENERATE_IP_ADDRESS_CLASS_PRIVATE_B"
	IpV4Class_PrivateC  IpV4Class = "GENERATE_IP_ADDRESS_CLASS_PRIVATE_C"
	IpV4Class_LinkLocal IpV4Class = "GENERATE_IP_ADDRESS_CLASS_LINK_LOCAL"
	IpV4Class_Multicast IpV4Class = "GENERATE_IP_ADDRESS_CLASS_MULTICAST"
	IpV4Class_Loopback  IpV4Class = "GENERATE_IP_ADDRESS_CLASS_LOOPBACK"
)

// Defined here -> https://www.meridianoutpost.com/resources/articles/IP-classes.php
// And here -> https://www.techtarget.com/whatis/definition/private-IP-addresshttps://www.techtarget.com/whatis/definition/private-IP-address
var ipv4Ranges = map[IpV4Class]struct {
	start net.IP
	end   net.IP
}{
	IpV4Class_PrivateA:  {net.ParseIP("10.0.0.0"), net.ParseIP("10.255.255.255")},
	IpV4Class_PrivateB:  {net.ParseIP("172.16.0.0"), net.ParseIP("172.31.255.255")},
	IpV4Class_PrivateC:  {net.ParseIP("192.168.0.0"), net.ParseIP("192.168.255.255")},
	IpV4Class_LinkLocal: {net.ParseIP("169.254.0.0"), net.ParseIP("169.254.255.255")},
	IpV4Class_Multicast: {net.ParseIP("224.0.0.0"), net.ParseIP("239.255.255.255")},
	IpV4Class_Loopback:  {net.ParseIP("127.0.0.0"), net.ParseIP("127.255.255.255")},
}

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Generates IPv4 or IPv6 addresses with support for different network classes.").
		Param(bloblang.NewInt64Param("max_length").Default(100000).Description("Specifies the maximum length for the generated data. This field ensures that the output does not exceed a certain number of characters.")).
		Param(bloblang.NewStringParam("version").Default(string(IpVersion_V4)).Description("IP version to generate: 'ipv4' or 'ipv6'")).
		Param(bloblang.NewStringParam("class").Default(string(IpV4Class_Public)).Description("IP class: 'public', 'private-a', 'private-b', 'private-c', 'link_local', 'multicast', 'loopback'")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("Optional seed for deterministic generation"))

	err := bloblang.RegisterFunctionV2("generate_ip", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		maxLength, err := args.GetInt64("max_length")
		if err != nil {
			return nil, err
		}
		version, err := args.GetString("version")
		if err != nil {
			return nil, err
		}

		class, err := args.GetString("class")
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

		versionStr := IpVersion(version)
		classStr := IpV4Class(class)

		return func() (any, error) {
			return generateIpAddress(randomizer, versionStr, classStr, maxLength)
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

func NewGenerateIpAddressOptsFromConfig(config *mgmtv1alpha1.GenerateIpAddress, maxlength *int64) (*GenerateIpAddressOpts, error) {
	if config == nil {
		return NewGenerateIpAddressOpts(nil, nil, nil, nil)
	}

	var version, class *string

	defaultVersion := string(IpVersion_V4)
	version = &defaultVersion
	if config.Version != nil {
		v := config.Version.String()
		version = &v
	}

	defaultClass := string(IpV4Class_Public)
	class = &defaultClass
	if config.Class != nil {
		c := config.Class.String()
		class = &c
	}

	return NewGenerateIpAddressOpts(maxlength, version, class, nil)
}

func (t *GenerateIpAddress) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateIpAddressOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}
	return generateIpAddress(parsedOpts.randomizer, IpVersion(parsedOpts.version), IpV4Class(parsedOpts.class), parsedOpts.maxLength)
}

func generateIpAddress(randomizer rng.Rand, version IpVersion, class IpV4Class, maxLength int64) (string, error) {
	var ip string
	var err error

	switch version {
	case IpVersion_V4:
		ip, err = generateIPv4Address(randomizer, class)
	case IpVersion_V6:
		ip = generateIPv6Address(randomizer)
	default:
		return "", fmt.Errorf("unsupported IP version: %s", version)
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

func generateIPv4Address(randomizer rng.Rand, class IpV4Class) (string, error) {
	if class == IpV4Class_Public {
		return generatePublicIPv4(randomizer)
	}

	ipRange, exists := ipv4Ranges[class]
	if !exists {
		return "", fmt.Errorf("unsupported IPv4 class: %s", class)
	}

	return generateIPInRange(randomizer, ipRange.start, ipRange.end), nil
}

func generateIPv6Address(randomizer rng.Rand) string {
	// Generate regular IPv6 address
	groups := make([]string, 8)
	for i := 0; i < 8; i++ {
		groups[i] = fmt.Sprintf("%04x", randomizer.Intn(65536))
	}
	return strings.Join(groups, ":")
}

func generatePublicIPv4(randomizer rng.Rand) (string, error) {
	maxAttempts := 1000 // Prevent infinite loop
	for i := 0; i < maxAttempts; i++ {
		ip := make(net.IP, 4)
		for i := range ip {
			ip[i] = byte(randomizer.Intn(256))
		}

		// Check if the IP falls within any reserved range
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
	// Also check for experimental/reserved range (240.0.0.0/4)
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

func inRange(ip, start, end net.IP) bool {
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

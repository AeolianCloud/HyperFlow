package ippool

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

const maxAddressesPerBatch = 254

type IPRange struct {
	Start net.IP
	End   net.IP
}

func ParseIPRange(s string) (IPRange, error) {
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return IPRange{}, fmt.Errorf("invalid IP range format: %s", s)
	}
	start := net.ParseIP(strings.TrimSpace(parts[0]))
	end := net.ParseIP(strings.TrimSpace(parts[1]))
	if start == nil || end == nil || start.To4() == nil || end.To4() == nil {
		return IPRange{}, fmt.Errorf("invalid IP address in range: %s", s)
	}
	if ipToInt(start) > ipToInt(end) {
		return IPRange{}, fmt.Errorf("start IP must be <= end IP: %s", s)
	}
	return IPRange{Start: start, End: end}, nil
}

func ExpandIPRange(r IPRange) ([]string, error) {
	start := ipToInt(r.Start)
	end := ipToInt(r.End)
	count := int(end - start + 1)
	if count > maxAddressesPerBatch {
		return nil, fmt.Errorf("IP range exceeds maximum of %d addresses (got %d)", maxAddressesPerBatch, count)
	}
	if count <= 0 {
		return nil, fmt.Errorf("invalid IP range")
	}
	ips := make([]string, 0, count)
	for i := start; i <= end; i++ {
		ips = append(ips, intToIP(i).String())
	}
	return ips, nil
}

func ParseAndExpandAddresses(inputs []string) ([]string, error) {
	seen := make(map[string]bool)
	var result []string

	for _, input := range inputs {
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		var ips []string
		if strings.Contains(input, "-") {
			r, err := ParseIPRange(input)
			if err != nil {
				return nil, err
			}
			ips, err = ExpandIPRange(r)
			if err != nil {
				return nil, err
			}
		} else {
			ip := net.ParseIP(input)
			if ip == nil || ip.To4() == nil {
				return nil, fmt.Errorf("invalid IP address: %s", input)
			}
			ips = []string{input}
		}

		for _, ip := range ips {
			if seen[ip] {
				return nil, fmt.Errorf("duplicate IP address: %s", ip)
			}
			seen[ip] = true
			result = append(result, ip)
		}
	}

	if len(result) > maxAddressesPerBatch {
		return nil, fmt.Errorf("total addresses exceed maximum of %d", maxAddressesPerBatch)
	}
	return result, nil
}

func ipToInt(ip net.IP) uint32 {
	ip = ip.To4()
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

func intToIP(n uint32) net.IP {
	return net.IPv4(byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}

func ParseCIDR(cidr string) (int, error) {
	cidr = strings.TrimSpace(cidr)
	n, err := strconv.Atoi(cidr)
	if err != nil {
		return 0, fmt.Errorf("invalid netmask: %s", cidr)
	}
	if n < 0 || n > 32 {
		return 0, fmt.Errorf("netmask must be between 0 and 32: %d", n)
	}
	return n, nil
}

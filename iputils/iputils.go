package iputils

import "net"

// IPType represents the type of an IP address
type IPType int

const (
	IPUnknown IPType = iota
	IPv4
	IPv6
)

// String returns a human-readable representation of the IPType
func (t IPType) String() string {
	switch t {
	case IPv4:
		return "v4"
	case IPv6:
		return "v6"
	default:
		return "unknown"
	}
}

// IsIPAddress returns true if the input string is a valid IPv4 or IPv6 address.
func IsIPAddress(s string) bool {
	return net.ParseIP(s) != nil
}

// IPVersion determines if the given string is an IPv4 or IPv6 address.
// Returns IPUnknown if the string is not a valid IP.
func IPVersion(s string) IPType {
	ip := net.ParseIP(s)
	if ip == nil {
		return IPUnknown
	}
	if ip.To4() != nil {
		return IPv4
	}
	return IPv6
}

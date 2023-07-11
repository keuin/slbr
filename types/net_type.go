package types

import "fmt"

type IpNetType string

var (
	IPv6Net IpNetType = "ipv6"
	IPv4Net IpNetType = "ipv4"
	IP64    IpNetType = "any"
)

// GetDialNetString returns the string accepted by net.Dialer::DialContext
func (t IpNetType) GetDialNetString() string {
	switch t {
	case IPv4Net:
		return "tcp4"
	case IPv6Net:
		return "tcp6"
	case IP64:
		return "tcp"
	}
	return ""
}

func (t IpNetType) String() string {
	return fmt.Sprintf("%s(%s)", string(t), t.GetDialNetString())
}

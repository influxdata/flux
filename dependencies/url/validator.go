package url

import (
	"fmt"
	"net"
	"net/url"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

// Validator reports whether a given URL is valid.
//
// XXX: `sql.validateDataSource` bypasses this for BigQuery DSNs since they have
// no host information in them and therefore can't have their IP validated.
// If validation is refactored to later consider more than just IPs, give the
// `sql.validateDataSource` another look.
type Validator interface {
	Validate(*url.URL) error
	ValidateIP(net.IP) error
}

// PassValidator passes all URLs
type PassValidator struct{}

func (PassValidator) Validate(*url.URL) error {
	return nil
}

func (PassValidator) ValidateIP(net.IP) error {
	return nil
}

// PrivateIPValidator validates that a url does not communicate with a private IP range
type PrivateIPValidator struct{}

func (v PrivateIPValidator) Validate(u *url.URL) error {
	ips, err := net.LookupIP(u.Hostname())
	if err != nil {
		return err
	}
	for _, ip := range ips {
		err = v.ValidateIP(ip)
		if err != nil {
			return err
		}
	}
	return nil
}

func (PrivateIPValidator) ValidateIP(ip net.IP) error {
	if isPrivateIP(ip) {
		// Intentionally return a vague message that we cannot connect to the host.
		// Do not explain why.
		return errors.New(codes.Invalid, "no such host")
	}
	return nil
}

// privateIPBlocks is a list of IP ranges that are defined as private.
var privateIPBlocks []*net.IPNet

func init() {
	for _, cidr := range []string{
		// IPv4 Special-Purpose Address Space
		// Address ranges taken from https://www.iana.org/assignments/iana-ipv4-special-registry/
		// that have the "Globally Reachable" flag marked as "False".
		"0.0.0.0/8",      // "This network" [RFC791], Section 3.2
		"10.0.0.0/8",     // Private-Use [RFC1918]
		"100.64.0.0/10",  // Shared Address Space [RFC6598]
		"127.0.0.0/8",    // Loopback [RFC1122], Section 3.2.1.3
		"169.254.0.0/16", // Link Local [RFC3927]
		"172.16.0.0/12",  // Private-Use [RFC1918]
		// The 192.0.0.0/24 block does include the addresses 192.0.0.9 &
		// 192.0.0.10, these are marked as "Globally Reachable" but are
		// for such specific protocols that blocking them will not
		// affect flux scripts.
		"192.0.0.0/24",       // IETF Protocol Assignments [RFC6890], Section 2.1
		"192.0.2.0/24",       // Documentation (TEST-NET-1) [RFC5737]
		"192.88.99.2/32",     // 6a44-relay anycast address [RFC6751]
		"192.168.0.0/16",     // Private-Use [RFC1918]
		"198.18.0.0/15",      // Benchmarking [RFC2544]
		"198.51.100.0/24",    // Documentation (TEST-NET-2) [RFC5737]
		"203.0.113.0/24",     // Documentation (TEST-NET-3) [RFC5737]
		"240.0.0.0/4",        // Reserved [RFC1112], Section 4
		"255.255.255.255/32", // Limited Broadcast [RFC8190] [RFC919], Section 7

		// IPv6 Special-Purpose Address Space
		// Address ranges taken from https://www.iana.org/assignments/iana-ipv6-special-registry/
		// that have the "Globally Reachable" flag marked as "False".
		"::1/128", // Loopback Address [RFC4291]
		"::/128",  // Unspecified Address [RFC4291]
		// The IPv4-mapped Address block is marked as not being globally
		// reachable, but is also how Go stores IPv4 Addresses, so
		// adding the range causes all IPv4 addresses to be blocked.
		// "::ffff:0:0/96", IPv4-mapped Address [RFC4291]
		"64:ff9b:1::/48", // IPv4-IPv6 Translat. [RFC8215]
		"100::/64",       // Discard-Only Address Block [RFC6666]
		"100:0:0:1::/64", // Dummy IPv6 Prefix [RFC9780]
		// The 2001::/23 block includes a number of ranges which are
		// marked as "Globally Reachable" but are for such specific
		// protocols that blocking them will not affect flux scripts.
		"2001::/23",     // IETF Protocol Assignments [RFC2928]
		"2001:db8::/32", // Documentation [RFC3849]
		"2002::/16",     // 6to4 [RFC3056]
		"3fff::/20",     // Documentation [RFC9637]
		"5f00::/16",     // Segment Routing (SRv6) SIDs [RFC9602]
		"fc00::/7",      // Unique-Local [RFC4193] [RFC8190]
		"fe80::/10",     // Link-Local Unicast [RFC4291]
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Errorf("parse error on %q: %v", cidr, err))
		}
		privateIPBlocks = append(privateIPBlocks, block)
	}
}

// isPrivateIP reports whether an IP exists in a known private IP space.
func isPrivateIP(ip net.IP) bool {
	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

type ErrorValidator struct{}

func (ErrorValidator) Validate(*url.URL) error {
	return errors.New(codes.Invalid, "Validator.Validate called on an error dependency")
}

func (ErrorValidator) ValidateIP(net.IP) error {
	return errors.New(codes.Invalid, "Validator.ValidateIP called on an error dependency")
}

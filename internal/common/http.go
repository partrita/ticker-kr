package common

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// DefaultUserAgent is the User-Agent header sent with every outbound quote
// request. The quote endpoints reject empty or non-browser agents, so the value
// stays browser-like; it lives here once so all clients send the same header.
const DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36"

// internalHostSuffixes are host names that always point inside the local network
// or the machine itself rather than at a public service.
var internalHostSuffixes = []string{
	"localhost",
	".localhost",
	".local",
	".internal",
	".localdomain",
}

// internalIPv4Ranges are IPv4 ranges that are not reachable as a public service:
// unspecified, private, carrier grade NAT, loopback, link-local, documentation,
// benchmarking and reserved space.
var internalIPv4Ranges = []string{
	"0.0.0.0/8",
	"10.0.0.0/8",
	"100.64.0.0/10",
	"127.0.0.0/8",
	"169.254.0.0/16",
	"172.16.0.0/12",
	"192.0.0.0/24",
	"192.0.2.0/24",
	"192.168.0.0/16",
	"198.18.0.0/15",
	"198.51.100.0/24",
	"203.0.113.0/24",
	"240.0.0.0/4",
}

// ValidateHTTPURL parses raw into an absolute http or https URL that is safe to
// request from.
//
// It rejects schemes other than http and https, URLs without a host, URLs that
// carry credentials, and hosts that address an internal network (loopback,
// private, link-local and other non-routable ranges). Scheme checks alone do not
// stop a configurable or redirected URL from reaching services inside the
// network, which is why the host is checked as well.
//
// allowLoopback relaxes the host check for loopback addresses only, so a client
// can talk to a server on the local machine; tests and local proxies need this.
// Private and link-local ranges stay blocked either way, and the flag must stay
// false for URLs that come from configuration.
//
// A DNS name is not resolved here: the check covers address literals and the
// usual local-only suffixes. Requests are checked again on every redirect, which
// is where a name could otherwise be swapped for an internal address.
func ValidateHTTPURL(raw string, allowLoopback bool) (*url.URL, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid URL %q: %w", raw, err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("invalid URL scheme: must be http or https")
	}

	if parsed.User != nil {
		return nil, fmt.Errorf("invalid URL %q: credentials in the URL are not allowed", raw)
	}

	host := parsed.Hostname()
	if host == "" {
		return nil, fmt.Errorf("invalid URL %q: host is required", raw)
	}

	if isInternalHost(host) && !(allowLoopback && isLoopbackHost(host)) {
		return nil, fmt.Errorf("invalid URL %q: host %q is not a public address", raw, host)
	}

	return parsed, nil
}

// isInternalHost reports whether host names a machine inside the local network
// or a non-routable range.
func isInternalHost(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(host, "."))

	if ip := net.ParseIP(host); ip != nil {
		return !isPublicIP(ip)
	}

	for _, suffix := range internalHostSuffixes {
		if host == suffix || strings.HasSuffix(host, suffix) {
			return true
		}
	}

	// A regular DNS name is allowed: it is expected to resolve to the public
	// service the user configured. Address literals above cover the case of a
	// URL that names an internal address directly.
	return false
}

// isLoopbackHost reports whether host addresses the local machine itself.
func isLoopbackHost(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(host, "."))

	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}

	return host == "localhost" || strings.HasSuffix(host, ".localhost")
}

// isPublicIP reports whether ip is an address that can be reached on the public
// internet.
func isPublicIP(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		ip = ip4
	} else {
		switch {
		case ip.IsLoopback(), ip.IsUnspecified(), ip.IsMulticast(),
			ip.IsLinkLocalUnicast(), ip.IsLinkLocalMulticast(),
			ip.IsInterfaceLocalMulticast(), ip.IsPrivate():
			return false
		}
	}

	for _, rawRange := range internalIPv4Ranges {
		_, ipRange, err := net.ParseCIDR(rawRange)
		if err != nil {
			continue
		}

		if ipRange.Contains(ip) {
			return false
		}
	}

	return true
}

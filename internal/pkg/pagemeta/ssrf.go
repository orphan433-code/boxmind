package pagemeta

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

func validateTargetURL(rawURL string) (*url.URL, error) {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid url")
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("unsupported url scheme")
	}

	if parsed.Hostname() == "" {
		return nil, fmt.Errorf("invalid url host")
	}

	if err := rejectPrivateHost(parsed.Hostname()); err != nil {
		return nil, err
	}

	return parsed, nil
}

func rejectPrivateHost(host string) error {
	host = strings.ToLower(strings.TrimSuffix(host, "."))

	if host == "localhost" {
		return fmt.Errorf("fetching localhost is not allowed")
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return nil
	}

	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return fmt.Errorf("fetching private addresses is not allowed")
	}

	return nil
}

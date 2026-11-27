package common_test

import (
	"testing"

	c "github.com/achannarasappa/ticker/v5/internal/common"
)

func TestValidateHTTPURLRejectsInternalTargets(t *testing.T) {
	rejected := []string{
		"http://localhost/api/v3/brokerage/market/products",
		"http://localhost:8080/v7/finance/quote",
		"http://service.internal/",
		"http://printer.local/",
		"http://127.0.0.1/",
		"http://127.1.2.3:9000/",
		"http://10.0.0.5/",
		"http://172.16.4.1/",
		"http://192.168.0.10/",
		"http://169.254.169.254/latest/meta-data/",
		"http://100.64.12.3/",
		"http://0.0.0.0/",
		"http://[::1]/",
		"http://[fd00::1]/",
		"http://[fe80::1]/",
		"http://[::ffff:10.0.0.5]/",
		"http://user:password@api.coinbase.com/",
		"ftp://api.coinbase.com/",
		"file:///etc/passwd",
		"ws://ws-feed.exchange.coinbase.com/",
		"/api/v3/brokerage/market/products",
		"api.coinbase.com",
	}

	for _, raw := range rejected {
		if _, err := c.ValidateHTTPURL(raw, false); err == nil {
			t.Errorf("ValidateHTTPURL(%q) should have been rejected", raw)
		}
	}
}

func TestValidateHTTPURLAcceptsPublicTargets(t *testing.T) {
	accepted := []string{
		"https://api.coinbase.com",
		"https://api.coinbase.com/api/v3/brokerage/market/products",
		"https://query1.finance.yahoo.com/v7/finance/quote",
		"http://93.184.216.34/",
		"https://[2606:4700:4700::1111]/",
		"https://api.coinbase.com:8443",
	}

	for _, raw := range accepted {
		parsed, err := c.ValidateHTTPURL(raw, false)
		if err != nil {
			t.Errorf("ValidateHTTPURL(%q) should have been accepted: %v", raw, err)
			continue
		}

		if parsed == nil || parsed.Host == "" {
			t.Errorf("ValidateHTTPURL(%q) returned an empty URL", raw)
		}
	}
}

func TestValidateHTTPURLAllowsLoopbackOnlyWhenOptedIn(t *testing.T) {
	accepted := []string{
		"http://127.0.0.1:40123/v7/finance/quote",
		"http://localhost:40123/v7/finance/quote",
		"http://[::1]:40123/v7/finance/quote",
	}

	for _, raw := range accepted {
		if _, err := c.ValidateHTTPURL(raw, true); err != nil {
			t.Errorf("ValidateHTTPURL(%q) should allow loopback when opted in: %v", raw, err)
		}
	}

	// Opting in to loopback must not open the rest of the internal network.
	rejected := []string{
		"http://169.254.169.254/latest/meta-data/",
		"http://10.0.0.5/",
		"http://192.168.0.10/",
	}

	for _, raw := range rejected {
		if _, err := c.ValidateHTTPURL(raw, true); err == nil {
			t.Errorf("ValidateHTTPURL(%q) should stay rejected even when loopback is allowed", raw)
		}
	}
}

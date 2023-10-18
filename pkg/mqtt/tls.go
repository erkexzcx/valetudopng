package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
)

func newTLSConfig(caFile string, insecureSkipVerify bool, minTLSVersion string) (tlsConfig *tls.Config, err error) {
	config := &tls.Config{}

	// Add CA file if provided to CA store
	if caFile != "" {
		certpool := x509.NewCertPool()
		pemCerts, err := os.ReadFile(caFile)
		if err != nil {
			return nil, err
		}
		certpool.AppendCertsFromPEM(pemCerts)
		config.RootCAs = certpool
	}

	// if 'true', then TLS verification is skipped
	config.InsecureSkipVerify = insecureSkipVerify

	// Set min TLS version
	switch minTLSVersion {
	case "": // Do nothing - defaults to Go's default
	case "1.0":
		config.MinVersion = tls.VersionTLS10
	case "1.1":
		config.MinVersion = tls.VersionTLS11
	case "1.2":
		config.MinVersion = tls.VersionTLS12
	case "1.3":
		config.MinVersion = tls.VersionTLS13
	default:
		return nil, errors.New("unrecognized TLS version " + minTLSVersion)
	}

	return config, nil
}

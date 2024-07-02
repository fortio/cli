// Fortio CLI/Main utilities.
//
// (c) 2024 Fortio Authors
// See LICENSE

//go:build !no_tls_fallback && !no_net
// +build !no_tls_fallback,!no_net

package cli // import "fortio.org/cli"

// golang.org/x/crypto/x509roots/fallback blank import below is because this is a base for all our main package,
// the CA bundle is needed for FROM scratch images to work with outcalls to internet valid TLS certs (https).
// See https://github.com/fortio/multicurl/pull/146 for instance.
import _ "golang.org/x/crypto/x509roots/fallback" // This is a base for main, see extended comment above.

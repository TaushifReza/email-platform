// Package smtpclient is the SMTP transport used to hand off one email to a
// provider. TLS is mandatory (STARTTLS or implicit), certificates and
// hostnames are verified, the minimum is TLS 1.2, and verification is never
// disabled. Plaintext SMTP is refused. The SSRF-safe dialer is injected by
// the caller and is the only network path.
package smtpclient

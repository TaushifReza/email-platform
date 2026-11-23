// Package dialer provides an SSRF-safe network dialer for user-supplied SMTP
// provider addresses. It classifies IP addresses, restricts ports to the
// allowed set, and re-checks the concrete address inside the dial Control
// function to defeat DNS rebinding.
package dialer

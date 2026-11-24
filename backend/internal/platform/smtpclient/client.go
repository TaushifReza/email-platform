package smtpclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"
)

var (
	ErrInvalidConfig = errors.New("smtpclient: invalid configuration")
	ErrPlaintext     = errors.New("smtpclient: plaintext SMTP is not allowed")
)

type TLSMode string

const (
	ModeSTARTTLS TLSMode = "starttls"
	ModeImplicit TLSMode = "implicit"
)

type Dialer interface {
	DialContext(ctx context.Context, network, addr string) (net.Conn, error)
}

type Result struct {
	Class    Class
	Stage    Stage
	Code     int
	Enhanced string
	Reply    string
	Err      error
}

type Envelope struct {
	From string
	To   []string
	Data []byte
}

type Client struct {
	opts Options
}

type Options struct {
	Host     string
	Port     int
	Mode     TLSMode
	Username string
	Password string
	Timeout  time.Duration
	Dialer   Dialer
	RootCAs  *x509.CertPool
}

func New(opts Options) (*Client, error) {
	if strings.TrimSpace(opts.Host) == "" {
		return nil, fmt.Errorf("%w: host required", ErrInvalidConfig)
	}
	if opts.Dialer == nil {
		return nil, fmt.Errorf("%w: dialer required", ErrInvalidConfig)
	}
	if opts.Mode != ModeSTARTTLS && opts.Mode != ModeImplicit {
		return nil, ErrPlaintext
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 2 * time.Minute
	}
	return &Client{opts: opts}, nil
}

func (c *Client) Send(ctx context.Context, env Envelope) Result {
	timeout := c.opts.Timeout
	res := Result{Stage: StageConnect}
	defer func() {
		if r := recover(); r != nil {
			res.Class = Uncertain
			res.Err = fmt.Errorf("smtpclient: panic: %v", r)
		}
	}()

	addr := fmt.Sprintf("%s:%d", c.opts.Host, c.opts.Port)

	setDeadline := func(conn net.Conn, d time.Duration) {
		_ = conn.SetDeadline(time.Now().Add(d)) //nolint:errcheck
	}

	conn, err := c.opts.Dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		res.Code = 0
		res.Err = err
		res.Class = classify(StageConnect, 0, err.Error())
		return res
	}
	defer conn.Close()
	setDeadline(conn, timeout)
	res.Stage = StageTLS

	if c.opts.Mode == ModeImplicit {
		tlsCfg := &tls.Config{
			ServerName: c.opts.Host, MinVersion: tls.VersionTLS12, RootCAs: c.opts.RootCAs,
		}
		tc := tls.Client(conn, tlsCfg)
		if err := tc.HandshakeContext(ctx); err != nil {
			res.Err = err
			res.Class = classify(StageTLS, 0, err.Error())
			return res
		}
		conn = tc
	}

	setDeadline(conn, timeout)
	sc, err := smtp.NewClient(conn, c.opts.Host)
	if err != nil {
		res.Err = err
		res.Class = TemporaryConnection
		return res
	}
	defer func() { _ = sc.Quit() }() //nolint:errcheck

	if c.opts.Mode == ModeSTARTTLS {
		res.Stage = StageTLS
		ok, _ := sc.Extension("STARTTLS")
		if !ok {
			res.Err = errors.New("smtpclient: server does not support STARTTLS")
			res.Class = TemporaryConnection
			return res
		}
		tlsCfg := &tls.Config{
			ServerName: c.opts.Host, MinVersion: tls.VersionTLS12, RootCAs: c.opts.RootCAs,
		}
		if err := sc.StartTLS(tlsCfg); err != nil {
			res.Err = err
			res.Class = classify(StageTLS, 0, err.Error())
			return res
		}
	}

	if c.opts.Username != "" {
		res.Stage = StageAuth
		auth := smtp.PlainAuth("", c.opts.Username, c.opts.Password, c.opts.Host)
		if err := sc.Auth(auth); err != nil {
			res.Err = err
			res.Reply = err.Error()
			res.Code = smtpErrCode(err)
			res.Class = classify(StageAuth, res.Code, res.Reply)
			return res
		}
	}

	res.Stage = StageMailFrom
	if err := sc.Mail(env.From); err != nil {
		res.Err = err
		res.Reply = err.Error()
		res.Code = smtpErrCode(err)
		res.Class = classify(StageMailFrom, res.Code, res.Reply)
		return res
	}

	res.Stage = StageRcpt
	for _, to := range env.To {
		if err := sc.Rcpt(to); err != nil {
			res.Err = err
			res.Reply = err.Error()
			res.Code = smtpErrCode(err)
			res.Class = classify(StageRcpt, res.Code, res.Reply)
			return res
		}
	}

	res.Stage = StageData
	w, err := sc.Data()
	if err != nil {
		res.Err = err
		res.Reply = err.Error()
		res.Code = smtpErrCode(err)
		res.Class = classify(StageData, res.Code, res.Reply)
		return res
	}

	res.Stage = StageBody
	if _, err := w.Write(env.Data); err != nil {
		res.Err = err
		res.Reply = err.Error()
		res.Code = smtpErrCode(err)
		res.Class = Uncertain
		_ = w.Close()
		return res
	}
	if err := w.Close(); err != nil {
		res.Err = err
		res.Reply = err.Error()
		res.Code = smtpErrCode(err)
		// Body was sent; any failure here is ambiguous.
		res.Class = Uncertain
		return res
	}

	res.Stage = StageReply
	res.Class = Accepted
	res.Code = 250
	return res
}

func smtpErrCode(err error) int {
	var te *textproto.Error
	if errors.As(err, &te) {
		return te.Code
	}
	return 0
}

package attacker

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"

	"github.com/nakabonne/ali/storage"
)

const (
	DefaultRate        = 50
	DefaultDuration    = 10 * time.Second
	DefaultTimeout     = 30 * time.Second
	DefaultMethod      = http.MethodGet
	DefaultWorkers     = 10
	DefaultMaxWorkers  = math.MaxUint64
	DefaultMaxBody     = int64(-1)
	DefaultConnections = 10000
)

var DefaultLocalAddr = net.IPAddr{IP: net.IPv4zero}

// Options provides optional settings to attack.
type Options struct {
	Rate        int
	Duration    time.Duration
	Timeout     time.Duration
	Method      string
	Body        []byte
	MaxBody     int64
	Header      http.Header
	Workers     uint64
	MaxWorkers  uint64
	KeepAlive   bool
	Connections int
	HTTP2       bool
	LocalAddr   net.IPAddr
	Buckets     []time.Duration
	Resolvers   []string

	InsecureSkipVerify bool
	CACertificatePool  *x509.CertPool
	TLSCertificates    []tls.Certificate

	Attacker backedAttacker
}

type Attacker interface {
	// Attack keeps the request running for the specified period of time.
	// Results are sent to the given channel as soon as they arrive.
	// When the attack is over, it gives back final statistics.
	// TODO: Use storage instead of metricsCh
	Attack(ctx context.Context, metricsCh chan *Metrics)

	// Rate gives back the rate set to itself.
	Rate() int
	// Rate gives back the duration set to itself.
	Duration() time.Duration
	// Rate gives back the method set to itself.
	Method() string
}

func NewAttacker(storage storage.Writer, target string, opts *Options) (Attacker, error) {
	if target == "" {
		return nil, fmt.Errorf("target is required")
	}
	if opts == nil {
		opts = &Options{}
	}
	if opts.Method == "" {
		opts.Method = DefaultMethod
	}
	if opts.Workers == 0 {
		opts.Workers = DefaultWorkers
	}
	if opts.MaxWorkers == 0 {
		opts.MaxWorkers = DefaultMaxWorkers
	}
	if opts.MaxBody == 0 {
		opts.MaxBody = DefaultMaxBody
	}
	if opts.Connections == 0 {
		opts.Connections = DefaultConnections
	}
	if opts.LocalAddr.IP == nil {
		opts.LocalAddr = DefaultLocalAddr
	}
	if len(opts.Resolvers) > 0 {
		net.DefaultResolver = NewResolver(opts.Resolvers)
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: opts.InsecureSkipVerify,
		Certificates:       opts.TLSCertificates,
		RootCAs:            opts.CACertificatePool,
	}
	tlsConfig.BuildNameToCertificate()

	if opts.Attacker == nil {
		opts.Attacker = vegeta.NewAttacker(
			vegeta.Timeout(opts.Timeout),
			vegeta.Workers(opts.Workers),
			vegeta.MaxWorkers(opts.MaxWorkers),
			vegeta.MaxBody(opts.MaxBody),
			vegeta.Connections(opts.Connections),
			vegeta.KeepAlive(opts.KeepAlive),
			vegeta.HTTP2(opts.HTTP2),
			vegeta.LocalAddr(opts.LocalAddr),
			vegeta.TLSConfig(tlsConfig),
		)
	}
	return &attacker{
		target:             target,
		rate:               opts.Rate,
		duration:           opts.Duration,
		timeout:            opts.Timeout,
		method:             opts.Method,
		body:               opts.Body,
		maxBody:            opts.MaxBody,
		header:             opts.Header,
		workers:            opts.Workers,
		maxWorkers:         opts.MaxWorkers,
		keepAlive:          opts.KeepAlive,
		connections:        opts.Connections,
		http2:              opts.HTTP2,
		localAddr:          opts.LocalAddr,
		buckets:            opts.Buckets,
		resolvers:          opts.Resolvers,
		insecureSkipVerify: opts.InsecureSkipVerify,
		caCertificatePool:  opts.CACertificatePool,
		tlsCertificates:    opts.TLSCertificates,
		attacker:           opts.Attacker,
		storage:            storage,
	}, nil
}

type backedAttacker interface {
	Attack(vegeta.Targeter, vegeta.Pacer, time.Duration, string) <-chan *vegeta.Result
	Stop()
}

type attacker struct {
	target             string
	rate               int
	duration           time.Duration
	timeout            time.Duration
	method             string
	body               []byte
	maxBody            int64
	header             http.Header
	workers            uint64
	maxWorkers         uint64
	keepAlive          bool
	connections        int
	http2              bool
	localAddr          net.IPAddr
	buckets            []time.Duration
	resolvers          []string
	insecureSkipVerify bool
	caCertificatePool  *x509.CertPool
	tlsCertificates    []tls.Certificate

	attacker backedAttacker
	storage  storage.Writer
}

func (a *attacker) Attack(ctx context.Context, metricsCh chan *Metrics) {
	rate := vegeta.Rate{Freq: a.rate, Per: time.Second}
	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: a.method,
		URL:    a.target,
		Body:   a.body,
		Header: a.header,
	})

	metrics := &vegeta.Metrics{}
	if len(a.buckets) > 0 {
		metrics.Histogram = &vegeta.Histogram{Buckets: a.buckets}
	}

	for res := range a.attacker.Attack(targeter, rate, a.duration, "main") {
		select {
		case <-ctx.Done():
			a.attacker.Stop()
			return
		default:
			metrics.Add(res)
			m := newMetrics(metrics)
			err := a.storage.Insert(&storage.Result{
				Code:      res.Code,
				Timestamp: res.Timestamp,
				Latency:   res.Latency,
				P50:       m.Latencies.P50,
				P90:       m.Latencies.P90,
				P95:       m.Latencies.P95,
				P99:       m.Latencies.P99,
			})
			if err != nil {
				log.Printf("failed to insert results")
				continue
			}
			metricsCh <- m
		}
	}
	metrics.Close()
	metricsCh <- newMetrics(metrics)
}

func (a *attacker) Rate() int {
	return a.rate
}

func (a *attacker) Duration() time.Duration {
	return a.duration
}

func (a *attacker) Method() string {
	return a.method
}

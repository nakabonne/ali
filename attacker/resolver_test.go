package attacker

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/miekg/dns"
)

const (
	testDomain       = "test.notadomain"
	DNSServerAddress = "127.0.0.1"
	message          = "Test Server"
)

func TestNewResolver(t *testing.T) {
	done := make(chan struct{}) // for ensuring ds.PacketConn is not nil

	// prepare custom DNS server
	dns.HandleFunc(".", handleRequest)
	ds := dns.Server{
		Addr:              DNSServerAddress + ":0",
		Net:               "udp",
		ReadTimeout:       time.Millisecond * time.Duration(2000),
		WriteTimeout:      time.Millisecond * time.Duration(2000),
		NotifyStartedFunc: func() { close(done) },
	}

	go func() {
		if err := ds.ListenAndServe(); err != nil {
			t.Logf("got error during dns ListenAndServe: %s", err)
		}
	}()
	defer func() {
		_ = ds.Shutdown()
	}()

	<-done

	net.DefaultResolver = NewResolver([]string{ds.PacketConn.LocalAddr().String()})

	// test server for name resolution
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, message)
	}))
	defer ts.Close()

	tsURL, _ := url.Parse(ts.URL)
	_, port, _ := net.SplitHostPort(tsURL.Host)
	tsURL.Host = net.JoinHostPort(testDomain, port)

	resp, err := http.Get(tsURL.String())
	if err != nil {
		t.Fatalf("failed resolver round trip: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read respose body: %v", err)
	}

	if strings.TrimSpace(string(body)) != message {
		t.Errorf("reponse body mismatch, expected: '%s', but got '%s'", message, body)
	}
}

func handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)

	m.Answer = []dns.RR{
		&dns.A{
			Hdr: dns.RR_Header{
				Name:   r.Question[0].Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    1,
			},
			A: net.ParseIP(DNSServerAddress),
		},
	}

	w.WriteMsg(m)
}

package gui

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/terminal/terminalapi"

	"github.com/nakabonne/ali/attacker"
)

func keybinds(ctx context.Context, cancel context.CancelFunc, dr *drawer) func(*terminalapi.Keyboard) {
	return func(k *terminalapi.Keyboard) {
		switch k.Key {
		case keyboard.KeyCtrlC: // Quit
			cancel()
		case keyboard.KeyEnter: // Attack
			attack(ctx, dr)
		}
	}
}

func attack(ctx context.Context, d *drawer) {
	if d.chartDrawing {
		return
	}
	target := d.widgets.urlInput.Read()
	if _, err := url.ParseRequestURI(target); err != nil {
		d.messageCh <- fmt.Sprintf("Bad URL: %v", err)
		return
	}
	opts, err := makeOptions(d.widgets)
	if err != nil {
		d.messageCh <- err.Error()
		return
	}
	requestNum := opts.Rate * int(opts.Duration/time.Second)

	// To pre-allocate, run redrawChart on a per-attack basis.
	go d.redrawChart(ctx, requestNum)
	go d.redrawGauge(ctx, requestNum)
	go func(ctx context.Context, d *drawer, t string, o attacker.Options) {
		metrics := attacker.Attack(ctx, t, d.chartCh, o)
		d.metricsCh <- metrics
		d.chartCh <- &attacker.Result{End: true}
		d.messageCh <- "Attack completed"
	}(ctx, d, target, opts)
}

// makeOptions gives back an options for attacker, with the input from UI.
func makeOptions(w *widgets) (attacker.Options, error) {
	var (
		rate     int
		duration time.Duration
		timeout  time.Duration
		method   string
		header   = make(http.Header)
		body     []byte
		err      error
	)

	if s := w.rateLimitInput.Read(); s == "" {
		rate = attacker.DefaultRate
	} else {
		rate, err = strconv.Atoi(s)
		if err != nil {
			return attacker.Options{}, fmt.Errorf("Given rate limit %q isn't integer: %w", s, err)
		}
	}

	if s := w.durationInput.Read(); s == "" {
		duration = attacker.DefaultDuration
	} else {
		duration, err = time.ParseDuration(s)
		if err != nil {
			return attacker.Options{}, fmt.Errorf("Unparseable duration %q: %w", s, err)
		}
	}

	if s := w.timeoutInput.Read(); s == "" {
		timeout = attacker.DefaultTimeout
	} else {
		timeout, err = time.ParseDuration(s)
		if err != nil {
			return attacker.Options{}, fmt.Errorf("Unparseable timeout %q: %w", s, err)
		}
	}

	if method = w.methodInput.Read(); method != "" {
		if !validateMethod(method) {
			return attacker.Options{}, fmt.Errorf("Given method %q isn't an HTTP request method", method)
		}
	}

	if s := w.headerInput.Read(); s != "" {
		parts := strings.SplitN(s, ":", 2)
		if len(parts) != 2 {
			return attacker.Options{}, fmt.Errorf("Given header %q has a wrong format", s)
		}
		key, val := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		if key == "" || val == "" {
			return attacker.Options{}, fmt.Errorf("Given header %q has a wrong format", s)
		}
		// Add key/value directly to the http.Header (map[string][]string).
		// http.Header.Add() canonicalizes keys but the vegeta API is used
		// to test systems that require case-sensitive headers.
		header[key] = append(header[key], val)
	}

	if f := w.bodyInput.Read(); f != "" {
		if b, err := ioutil.ReadFile(f); err != nil {
			return attacker.Options{}, fmt.Errorf("Unable to open %q: %w", f, err)
		} else {
			body = b
		}
	}

	return attacker.Options{
		Rate:     rate,
		Duration: duration,
		Timeout:  timeout,
		Method:   method,
		Body:     body,
		Header:   header,
	}, nil
}

func validateMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace:
		return true
	}
	return false
}

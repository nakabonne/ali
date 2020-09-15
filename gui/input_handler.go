package gui

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nakabonne/ali/attacker"
)

// makeOptions gives back an options for attacker, with the input from UI.
func makeOptions(d *drawer) (attacker.Options, error) {
	var (
		rate     int
		duration time.Duration
		method   string
		body     string
		header   = make(http.Header)
		err      error
	)

	if s := d.widgets.rateLimitInput.Read(); s == "" {
		rate = attacker.DefaultRate
	} else {
		rate, err = strconv.Atoi(s)
		if err != nil {
			return attacker.Options{}, fmt.Errorf("Given rate limit %q isn't integer: %w", s, err)
		}
	}

	if s := d.widgets.durationInput.Read(); s == "" {
		duration = attacker.DefaultDuration
	} else {
		duration, err = time.ParseDuration(s)
		if err != nil {
			return attacker.Options{}, fmt.Errorf("Unparseable duration %q: %w", s, err)
		}
	}

	if method = d.widgets.methodInput.Read(); method != "" {
		if !validateMethod(method) {
			return attacker.Options{}, fmt.Errorf("Given method %q isn't an HTTP request method", method)
		}
	}

	body = d.widgets.bodyInput.Read()

	if s := d.widgets.headerInput.Read(); s != "" {
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

	return attacker.Options{
		Rate:     rate,
		Duration: duration,
		Method:   method,
		Body:     []byte(body),
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

package admission

import (
	"fmt"
	"net"

	"github.com/maxcelant/sinkplot/internal/schema"
)

type validator struct {
	app *schema.App
}

var validStrategies = map[string]bool{
	"round-robin": true,
	"random":      true,
	"weighted":    true,
}

var validMatches = map[string]bool{
	"prefix": true,
	"exact":  true,
	"regex":  true,
}

var validMethods = map[string]bool{
	"GET":     true,
	"POST":    true,
	"PUT":     true,
	"DELETE":  true,
	"PATCH":   true,
	"HEAD":    true,
	"OPTIONS": true,
}

// Validate ensures that all necessary fields are set and are correct
func Validate(app *schema.App) error {
	d := validator{app}
	if err := d.validatePorts(); err != nil {
		return fmt.Errorf("port validation failed: %w", err)
	}
	if err := d.validateStrategy(); err != nil {
		return fmt.Errorf("sink strategy validation failed: %w", err)
	}
	if err := d.validatePrefix(); err != nil {
		return fmt.Errorf("route prefix validation failed: %w", err)
	}
	// Add localhost and hostnames as a possibility here
	if err := d.validateIP(); err != nil {
		return fmt.Errorf("upstream IP validation failed: %w", err)
	}
	if err := d.validateMethods(); err != nil {
		return fmt.Errorf("route methods validation failed: %w", err)
	}
	if err := d.validateListenerPorts(); err != nil {
		return fmt.Errorf("listener ports validation failed: %w", err)
	}
	if err := d.validateRouteSinks(); err != nil {
		return fmt.Errorf("route sink validation failed: %w", err)
	}
	if err := d.validateRoutePaths(); err != nil {
		return fmt.Errorf("route path validation failed: %w", err)
	}
	return nil
}

func (v validator) validatePorts() error {
	for _, s := range v.app.Sinks {
		for _, u := range s.Upstreams {
			if u.Port < 0 || u.Port > 65535 {
				return fmt.Errorf("invalid port value %d", u.Port)
			}
		}
	}
	return nil
}

func (v validator) validateStrategy() error {
	for _, s := range v.app.Sinks {
		// This should never happen if you are running the defaulter first
		if s.Strategy == nil || *s.Strategy == "" {
			return fmt.Errorf("invalid strategy, value was null or blank")
		}
		if !validStrategies[*s.Strategy] {
			return fmt.Errorf("invalid strategy %q", *s.Strategy)
		}
	}
	return nil
}

func (v validator) validatePrefix() error {
	for _, r := range v.app.Routes {
		if r.Match == nil || *r.Match == "" {
			return fmt.Errorf("invalid route match, value was null or blank")
		}
		if !validMatches[*r.Match] {
			return fmt.Errorf("invalid route path matcher %q", *r.Match)
		}
	}
	return nil
}

func (v validator) validateIP() error {
	for _, s := range v.app.Sinks {
		for _, u := range s.Upstreams {
			if u.Address == "" {
				return fmt.Errorf("upstream address cannot be empty in sink %q", s.Name)
			}
			if u.Address == "localhost" {
				continue
			}
			// TODO: Validate hostnames as well
			if ip := net.ParseIP(u.Address); ip == nil {
				return fmt.Errorf("invalid IP address %q in sink %q", u.Address, s.Name)
			}
		}
	}
	return nil
}

func (v validator) validateMethods() error {
	for _, r := range v.app.Routes {
		if r.Methods == nil {
			continue
		}
		for _, m := range *r.Methods {
			if !validMethods[m] {
				return fmt.Errorf("invalid HTTP method %q in route %q", m, r.Path)
			}
		}
	}
	return nil
}

func (v validator) validateListenerPorts() error {
	for _, port := range v.app.Listeners {
		if port < 1 || port > 65535 {
			return fmt.Errorf("invalid listener port %d", port)
		}
	}
	return nil
}

func (v validator) validateRouteSinks() error {
	sinkNames := make(map[string]bool)
	for _, s := range v.app.Sinks {
		sinkNames[s.Name] = true
	}
	for _, r := range v.app.Routes {
		if r.Sink == "" {
			return fmt.Errorf("route %q has no sink specified", r.Path)
		}
		if !sinkNames[r.Sink] {
			return fmt.Errorf("route %q references unknown sink %q", r.Path, r.Sink)
		}
	}
	return nil
}

func (v validator) validateRoutePaths() error {
	for _, r := range v.app.Routes {
		if r.Path == "" {
			return fmt.Errorf("route path cannot be empty")
		}
		if r.Path[0] != '/' {
			return fmt.Errorf("route path %q must start with /", r.Path)
		}
	}
	return nil
}

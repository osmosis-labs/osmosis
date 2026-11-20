package core

import (
	"fmt"
	"sort"
)

// Rule checks one message whose signer is a listed address. ctx is the
// adapter's context (sdk.Context in SDK adapters), msg the concrete message,
// signer the listed signer. Return nil to allow, an error to reject. Rules
// may read state but must never write it.
type Rule func(c *Config, ctx any, msg any, signer []byte) error

// Registry maps message type URLs to rules. A listed signer may only send
// message types that are registered here, and only if the rule allows it.
type Registry struct {
	rules map[string]Rule
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry { return &Registry{rules: map[string]Rule{}} }

// Register adds a rule. Registering the same type twice is a programming
// error and panics at startup.
func (r *Registry) Register(typeURL string, rule Rule) {
	if _, dup := r.rules[typeURL]; dup {
		panic(fmt.Sprintf("recovery lock: rule for %s registered twice", typeURL))
	}
	r.rules[typeURL] = rule
}

// Check applies the rule for typeURL. Unregistered types are rejected.
func (r *Registry) Check(c *Config, ctx any, typeURL string, msg any, signer []byte) error {
	rule, ok := r.rules[typeURL]
	if !ok {
		return fmt.Errorf("recovery lock: %s is not allowed for listed accounts", typeURL)
	}
	return rule(c, ctx, msg, signer)
}

// Types returns the registered type URLs in sorted order.
func (r *Registry) Types() []string {
	out := make([]string, 0, len(r.rules))
	for t := range r.rules {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

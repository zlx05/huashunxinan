package engine

import (
	"embed"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"gopkg.in/yaml.v3"

	"bannerfingerprint/internal/model"
)

//go:embed rules/fingerprints.yaml
var defaultRules embed.FS

const defaultRulesPath = "rules/fingerprints.yaml"

// unknownProtocol is reported when no rule matches a record.
const unknownProtocol = "unknown"

// RulesModel is the top-level shape of the fingerprints.yaml file.
type RulesModel struct {
	Version int               `yaml:"version"`
	Rules   []FingerprintRule `yaml:"rules"`
}

// Engine holds compiled fingerprint rules.
type Engine struct {
	rules []*FingerprintRule
}

// New loads the embedded default rule set and compiles it.
func New() (*Engine, error) {
	data, err := defaultRules.ReadFile(defaultRulesPath)
	if err != nil {
		return nil, fmt.Errorf("read embedded rules: %w", err)
	}
	return load(data)
}

// NewFromFile loads rules from an external YAML file, overriding the embedded set.
func NewFromFile(path string) (*Engine, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read rules file: %w", err)
	}
	return load(data)
}

func load(data []byte) (*Engine, error) {
	var m RulesModel
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse rules: %w", err)
	}
	e := &Engine{rules: make([]*FingerprintRule, 0, len(m.Rules))}
	for i := range m.Rules {
		r := &m.Rules[i]
		if err := r.compileRegexes(); err != nil {
			return nil, fmt.Errorf("rule %q: %w", r.Name, err)
		}
		e.rules = append(e.rules, r)
	}
	return e, nil
}

// Recognize computes the fingerprint for a single input record.
func (e *Engine) Recognize(in model.Input) model.Result {
	banner := normalizeBanner(in.Banner)
	res := model.Result{IP: in.IP, Port: in.Port}

	bestIdx, bestScore, bestContent := -1, -1, false
	for i, r := range e.rules {
		contentHit := r.matchCompiled.MatchString(banner)
		portHit := contains(r.Ports, in.Port)
		if !contentHit && !portHit {
			continue
		}

		score := 0
		if contentHit {
			score += 100
		}
		if portHit {
			score += 10
		}
		if extractProduct(r, banner) != "" {
			score += 5
		}
		if extractVersion(r, banner) != "" {
			score += 5
		}

		if score > bestScore || (bestIdx >= 0 && score == bestScore && r.Confidence > e.rules[bestIdx].Confidence) {
			bestIdx, bestScore, bestContent = i, score, contentHit
		}
	}

	if bestIdx < 0 {
		res.Protocol = unknownProtocol
		res.Confidence = 0.3
		return res
	}

	r := e.rules[bestIdx]
	res.Protocol = r.Protocol
	res.Product = extractProduct(r, banner)
	res.Version = extractVersion(r, banner)
	res.OSHint = extractOS(r, banner)
	if res.OSHint == "" {
		res.OSHint = detectOS(banner)
	}

	// Confidence: base scored down when matched only by port (weaker evidence).
	conf := r.Confidence
	if !bestContent {
		conf -= 0.2
	}
	if conf < 0 {
		conf = 0
	}
	if conf > 1 {
		conf = 1
	}
	res.Confidence = conf
	return res
}

// RecognizeBatch computes fingerprints for many records.
func (e *Engine) RecognizeBatch(inputs []model.Input) []model.Result {
	results := make([]model.Result, 0, len(inputs))
	for _, in := range inputs {
		results = append(results, e.Recognize(in))
	}
	return results
}

func extractProduct(r *FingerprintRule, banner string) string {
	if r.Product == nil {
		return ""
	}
	return r.Product.extract(banner)
}

func extractVersion(r *FingerprintRule, banner string) string {
	if r.Version == nil {
		return ""
	}
	return r.Version.extract(banner)
}

func extractOS(r *FingerprintRule, banner string) string {
	if r.OS == nil {
		return ""
	}
	return r.OS.extract(banner)
}

func contains(list []int, v int) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

var hexEscape = regexp.MustCompile(`\\x([0-9a-fA-F]{2})`)

// normalizeBanner converts literal "\xNN" text into the actual byte, so raw
// banners that arrive as escaped text (e.g. MySQL handshakes with NUL bytes)
// match reliably. Already-decoded control characters are left untouched.
func normalizeBanner(s string) string {
	return hexEscape.ReplaceAllStringFunc(s, func(m string) string {
		sub := hexEscape.FindStringSubmatch(m)
		v, err := strconv.ParseUint(sub[1], 16, 8)
		if err != nil {
			return m
		}
		return string([]byte{byte(v)})
	})
}

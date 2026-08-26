package engine

import "regexp"

// CaptureRule extracts a field from the banner: either by capturing a regex
// group, or by returning a fixed literal value.
type CaptureRule struct {
	Regex   string `yaml:"regex"`
	Group   int    `yaml:"group"`
	Literal string `yaml:"literal"`

	compiled *regexp.Regexp
}

// FingerprintRule is a single signature template for one protocol/product.
type FingerprintRule struct {
	Name       string       `yaml:"name"`
	Protocol   string       `yaml:"protocol"`
	Ports      []int        `yaml:"ports"`
	Match      string       `yaml:"match"`
	Product    *CaptureRule `yaml:"product"`
	Version    *CaptureRule `yaml:"version"`
	OS         *CaptureRule `yaml:"os"`
	Confidence float64      `yaml:"confidence"`

	matchCompiled *regexp.Regexp
}

// compileRegexes compiles the rule's regexes once so matching is cheap.
func (r *FingerprintRule) compileRegexes() error {
	re, err := regexp.Compile(r.Match)
	if err != nil {
		return err
	}
	r.matchCompiled = re

	for _, c := range []*CaptureRule{r.Product, r.Version, r.OS} {
		if c == nil || c.Literal != "" {
			continue
		}
		cre, err := regexp.Compile(c.Regex)
		if err != nil {
			return err
		}
		c.compiled = cre
	}
	return nil
}

// extract applies a CaptureRule to the banner, returning the captured value
// or the literal if set. Empty string when nothing matches.
func (c *CaptureRule) extract(banner string) string {
	if c == nil {
		return ""
	}
	if c.Literal != "" {
		return c.Literal
	}
	m := c.compiled.FindStringSubmatch(banner)
	if m == nil {
		return ""
	}
	// group defaults to 1 if not explicitly set but a group exists.
	g := c.Group
	if g <= 0 {
		g = 1
	}
	if g >= len(m) {
		return ""
	}
	return m[g]
}

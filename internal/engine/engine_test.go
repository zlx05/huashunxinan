package engine

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	"bannerfingerprint/internal/model"
)

// TestRecognizeBatchMatchesExpected verifies the engine reproduces the
// recognition depth shown in goal.txt for all bundled sample records.
func TestRecognizeBatchMatchesExpected(t *testing.T) {
	e, err := New()
	if err != nil {
		t.Fatalf("init engine: %v", err)
	}

	inputs := loadInput(t, "../../testdata/input.json")
	expected := loadExpected(t, "../../testdata/expected.json")
	got := e.RecognizeBatch(inputs)

	if len(got) != len(expected) {
		t.Fatalf("got %d results, want %d", len(got), len(expected))
	}

	for i := range got {
		want, have := expected[i], got[i]
		if have.IP != want.IP || have.Port != want.Port {
			t.Errorf("record %d: ip/port mismatch: got %s/%d want %s/%d", i, have.IP, have.Port, want.IP, want.Port)
		}
		if have.Protocol != want.Protocol {
			t.Errorf("record %s/%d protocol: got %q want %q", have.IP, have.Port, have.Protocol, want.Protocol)
		}
		if have.Product != want.Product {
			t.Errorf("record %s/%d product: got %q want %q", have.IP, have.Port, have.Product, want.Product)
		}
		if have.Version != want.Version {
			t.Errorf("record %s/%d version: got %q want %q", have.IP, have.Port, have.Version, want.Version)
		}
		if have.OSHint != want.OSHint {
			t.Errorf("record %s/%d os_hint: got %q want %q", have.IP, have.Port, have.OSHint, want.OSHint)
		}
		if math.Abs(have.Confidence-want.Confidence) > 0.01 {
			t.Errorf("record %s/%d confidence: got %.2f want %.2f", have.IP, have.Port, have.Confidence, want.Confidence)
		}
	}
}

// TestNormalizeBanner confirms "\xNN" text is converted to the real byte.
func TestNormalizeBanner(t *testing.T) {
	cases := map[string]string{
		// `\x00` literal text -> NUL byte; `\n` (already-decoded newline) left as-is.
		"J\\x00\\x00\\x00\n8.0.32\\x00": string([]byte{'J', 0, 0, 0, '\n', '8', '.', '0', '.', '3', '2', 0}),
		"plain banner":                    "plain banner",
		"\\x16\\x03":                      string([]byte{0x16, 0x03}),
	}
	for in, want := range cases {
		if got := normalizeBanner(in); got != want {
			t.Errorf("normalizeBanner(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestEngineEdgeCases covers individual protocol entry points not covered by
// the full expected dataset.
func TestEngineEdgeCases(t *testing.T) {
	e, err := New()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		in       model.Input
		protocol string
		product  string
		version  string
	}{
		{"ssh-no-os", model.Input{Port: 22, Banner: "SSH-2.0-OpenSSH_7.2"}, "SSH", "OpenSSH", "7.2"},
		{"http-no-server", model.Input{Port: 80, Banner: "HTTP/1.1 200 OK"}, "HTTP", "", ""},
		{"mysql-port-only", model.Input{Port: 3306, Banner: "garbage"}, "MySQL", "MySQL", ""},
		{"redis-noauth", model.Input{Port: 6379, Banner: "-NOAUTH Authentication required."}, "Redis", "Redis", ""},
		{"unknown", model.Input{Port: 12345, Banner: "QUIT"}, unknownProtocol, "", ""},
	}

	for _, tc := range tests {
		r := e.Recognize(tc.in)
		if r.Protocol != tc.protocol {
			t.Errorf("%s: protocol got %q want %q", tc.name, r.Protocol, tc.protocol)
		}
		if r.Product != tc.product {
			t.Errorf("%s: product got %q want %q", tc.name, r.Product, tc.product)
		}
		if r.Version != tc.version {
			t.Errorf("%s: version got %q want %q", tc.name, r.Version, tc.version)
		}
	}
}

func loadInput(t *testing.T, path string) []model.Input {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var in []model.Input
	if err := json.Unmarshal(data, &in); err != nil {
		t.Fatal(err)
	}
	return in
}

func loadExpected(t *testing.T, path string) []model.Result {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var exp []model.Result
	if err := json.Unmarshal(data, &exp); err != nil {
		t.Fatal(err)
	}
	return exp
}

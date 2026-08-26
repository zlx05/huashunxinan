package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"bannerfingerprint/internal/model"
)

func main() {
	input := flag.String("input", "", "input JSON file path (default: read stdin)")
	server := flag.String("server", "http://localhost:8080", "server base URL")
	output := flag.String("output", "", "optional output JSON file; default prints to stdout")
	timeout := flag.Duration("timeout", 30*time.Second, "request timeout")
	flag.Parse()

	inputs, err := readInputs(*input)
	if err != nil {
		fatal(err)
	}

	results, err := post(*server, inputs, *timeout)
	if err != nil {
		fatal(err)
	}

	out, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fatal(err)
	}

	if *output != "" {
		if err := os.WriteFile(*output, out, 0o644); err != nil {
			fatal(err)
		}
	}
	fmt.Println(string(out))
}

func readInputs(path string) ([]model.Input, error) {
	var r io.Reader = os.Stdin
	if path != "" {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		r = f
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var inputs []model.Input
	if err := json.Unmarshal(data, &inputs); err != nil {
		return nil, err
	}
	return inputs, nil
}

func post(server string, inputs []model.Input, timeout time.Duration) ([]model.Result, error) {
	body, err := json.Marshal(inputs)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Post(server+"/fingerprint", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(msg))
	}
	var results []model.Result
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}
	return results, nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "client:", err)
	os.Exit(1)
}

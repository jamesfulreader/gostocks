package config

import (
	"bufio"
	"os"
	"strings"
)

// LoadEnv reads a .env file in the project root (if present) and sets env vars.
// It ignores malformed lines and does not override variables already set.
func LoadEnv() error {
	f, err := os.Open(".env")
	if err != nil {
		return nil // .env optional
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// simple KEY=VALUE parser
		if i := strings.Index(line, "="); i > 0 {
			k := strings.TrimSpace(line[:i])
			v := strings.Trim(strings.TrimSpace(line[i+1:]), "\"")
			if _, exists := os.LookupEnv(k); !exists {
				_ = os.Setenv(k, v)
			}
		}
	}
	return nil
}

func GetenvDefault(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

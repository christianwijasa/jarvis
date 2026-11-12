package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const SecretsFileName = "secrets"

type secretsFile struct {
	Token string `yaml:"token"`
}

func SecretsPath(name string) (string, error) {
	dir, err := Dir(name)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, SecretsFileName), nil
}

// TokenSet is true when a secrets file with a non-empty token exists.
func TokenSet(name string) bool {
	tok, err := ReadToken(name)
	return err == nil && tok != ""
}

func ReadToken(name string) (string, error) {
	path, err := SecretsPath(name)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	var s secretsFile
	if err := yaml.Unmarshal(data, &s); err != nil {
		return "", fmt.Errorf("parse %s: %w", path, err)
	}
	return strings.TrimSpace(s.Token), nil
}

func WriteToken(name, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return DeleteToken(name)
	}
	dir, err := Dir(name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, SecretsFileName)
	var b strings.Builder
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	if err := enc.Encode(secretsFile{Token: token}); err != nil {
		return err
	}
	if err := enc.Close(); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o600); err != nil {
		return err
	}
	return os.Chmod(path, 0o600)
}

func DeleteToken(name string) error {
	path, err := SecretsPath(name)
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

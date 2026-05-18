package proxy

import (
	"encoding/json"
	"os"
	"path/filepath"
)

var keyFilePath = defaultKeyFilePath

func defaultKeyFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".claude-code-proxy/key.json"
	}
	return filepath.Join(home, ".claude-code-proxy", "key.json")
}

type keyFile struct {
	APIKey string `json:"api_key"`
}

func SaveKeyFile(apiKey string) error {
	path := keyFilePath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.Marshal(keyFile{APIKey: apiKey})
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func LoadKeyFile() (string, error) {
	path := keyFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var kf keyFile
	if err := json.Unmarshal(data, &kf); err != nil {
		return "", err
	}
	return kf.APIKey, nil
}

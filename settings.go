package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	settingsFile     = "settings.json"
	settingsFileMode = 0o600
	homePrefix       = "~"
)

var errNotFolder = errors.New("not a folder")

type Settings struct {
	Vault string `json:"vault"`
}

// LoadSettings reads the settings kept in dir, the extension's data
// directory. A first run has no file yet and answers the zero Settings.
func LoadSettings(dir string) (Settings, error) {
	var s Settings
	data, err := os.ReadFile(filepath.Join(dir, settingsFile))
	if errors.Is(err, fs.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	return s, json.Unmarshal(data, &s)
}

func SaveSettings(dir string, s Settings) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, settingsFile), data, settingsFileMode)
}

// ResolveVault turns what the person typed into the absolute path of an
// existing folder, reading a leading ~ as their home directory. A link is
// followed here because listing a vault does not follow one.
func ResolveVault(typed string) (string, error) {
	path := strings.TrimSpace(typed)
	if path == homePrefix || strings.HasPrefix(path, homePrefix+"/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[len(homePrefix):])
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", errNotFolder
	}
	return path, nil
}

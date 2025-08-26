package config

import (
	"os"
	"path/filepath"
)

// Paths holds important config paths.
type Paths struct {
	KnownHosts string
	CacheDB    string
	ConfigFile string
}

// ResolvePaths determines paths based on XDG (simplified).
func ResolvePaths() Paths {
	cfgHome := os.Getenv("XDG_CONFIG_HOME")
	if cfgHome == "" {
		cfgHome = filepath.Join(os.Getenv("HOME"), ".config")
	}
	cacheHome := os.Getenv("XDG_CACHE_HOME")
	if cacheHome == "" {
		cacheHome = filepath.Join(os.Getenv("HOME"), ".cache")
	}
	pDir := filepath.Join(cfgHome, "photoptim")
	_ = os.MkdirAll(pDir, 0o700)
	cDir := filepath.Join(cacheHome, "photoptim")
	_ = os.MkdirAll(cDir, 0o700)
	return Paths{
		KnownHosts: filepath.Join(pDir, "known_hosts"),
		CacheDB:    filepath.Join(cDir, "cache.db"),
		ConfigFile: filepath.Join(pDir, "config.yaml"),
	}
}

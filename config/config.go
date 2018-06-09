package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/reviewboard/rb-gateway/repositories"
)

const DefaultConfigPath = "config.json"

type repositoryData struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Scm  string `json:"scm"`
}

type Config struct {
	Port           uint16           `json:"port"`
	Username       string           `json:"username"`
	Password       string           `json:"password"`
	UseTLS         bool             `json:"useTLS"`
	SSLCertificate string           `json:"sslCertificate"`
	SSLKey         string           `json:"sslKey"`
	RepositoryData []repositoryData `json:"repositories"`

	Repositories map[string]repositories.Repository
}

func Load(path string) (*Config, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err = json.Unmarshal(content, &config); err != nil {
		return nil, err
	}

	var cfgDir string
	if cfgDir, err = filepath.Abs(path); err != nil {
		return nil, err
	} else {
		cfgDir = filepath.Dir(cfgDir)
	}

	if err = validate(cfgDir, &config); err != nil {
		return nil, err
	}

	config.Repositories = make(map[string]repositories.Repository)

	for _, repo := range config.RepositoryData {
		switch repo.Scm {
		case "git":
			config.Repositories[repo.Name] = &repositories.GitRepository{
				repositories.RepositoryInfo{
					Name: repo.Name,
					Path: repo.Path,
				},
			}

		default:
			log.Printf("Unknown SCM '%s' while loading configuration '%s'; ignoring.", repo.Scm, path)
		}
	}

	return &config, nil
}

func validate(cfgDir string, config *Config) (err error) {
	missingFields := []string{}

	if config.Port == 0 {
		log.Printf("WARNING: Port missing, defaulting to 8888.")
		config.Port = 8888
	}

	if config.Username == "" {
		missingFields = append(missingFields, "username")
	}

	if config.Password == "" {
		missingFields = append(missingFields, "password")
	}

	if len(config.RepositoryData) == 0 {
		missingFields = append(missingFields, "repositories")
	}

	if config.UseTLS {
		if config.SSLCertificate == "" {
			missingFields = append(missingFields, "ssl_certificate")
		} else {
			config.SSLCertificate = resolvePath(cfgDir, config.SSLCertificate)
		}

		if config.SSLKey == "" {
			missingFields = append(missingFields, "ssl_key")
		} else {
			config.SSLKey = resolvePath(cfgDir, config.SSLKey)
		}
	}

	if len(missingFields) != 0 {
		err = fmt.Errorf("Some required fields were missing from the configuration: %s.", strings.Join(missingFields, ","))
	}

	return
}

// Resolve a path so that . is treated as cfgDir
func resolvePath(cfgDir string, path string) string {
	if !filepath.IsAbs(path) {
		path = filepath.Join(cfgDir, path)
	}
	return path
}

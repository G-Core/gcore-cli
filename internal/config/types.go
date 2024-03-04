package config

import (
	"fmt"
	"github.com/AlekSi/pointer"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strconv"
)

const (
	DefaultProfile = "default"
	DefaultAPI     = "https://api.gcore.com"
)

const (
	EnvConfigPath          = "GCORE_CONFIG"
	EnvProfileURL          = "GCORE_API_URL"
	EnvProfileAPIKey       = "GCORE_API_KEY"
	EnvProfileCloudProject = "GCORE_CLOUD_PROJECT"
	EnvProfileCloudRegion  = "GCORE_CLOUD_REGION"
)

type Profile struct {
	ApiUrl       *string `yaml:"api-url,omitempty"       json:"api-url,omitempty"`
	ApiKey       *string `yaml:"api-key,omitempty"       json:"api-key,omitempty"`
	CloudProject *int    `yaml:"cloud-project,omitempty" json:"cloud-project,omitempty"`
	CloudRegion  *int    `yaml:"cloud-region,omitempty"  json:"cloud-region,omitempty"`
}

func (p *Profile) IsLocal() bool {
	if p.ApiUrl == nil {
		return false
	}

	if *p.ApiUrl == DefaultAPI {
		return false
	}

	return true
}

type Config struct {
	Profile       `yaml:",inline"`
	ActiveProfile string              `yaml:"profile"            json:"profile,omitempty"`
	Profiles      map[string]*Profile `yaml:"profiles,omitempty" json:"profiles,omitempty"`
}

func NewDefault() *Config {
	return &Config{
		ActiveProfile: DefaultProfile,
		Profile: Profile{
			ApiUrl: pointer.To(DefaultAPI),
		},
	}
}

func (c *Config) String() string {
	body, _ := yaml.Marshal(c)

	return string(body)
}

func (c *Config) Load(path string) error {
	body, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if err := yaml.Unmarshal(body, c); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

func (c *Config) Save(path string) error {
	body, err := yaml.Marshal(*c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	if err := os.WriteFile(path, body, 0644); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

func (c *Config) GetProfile(name string) (*Profile, error) {
	if name == DefaultProfile {
		return &c.Profile, nil
	}

	if c.Profiles == nil {
		return nil, fmt.Errorf("profile '%s' isn't exist", name)
	}

	p, exist := c.Profiles[name]
	if !exist {
		return nil, fmt.Errorf("profile '%s' isn't exist", name)
	}

	return MergeProfiles(&c.Profile, p), nil
}

func (c *Config) SetProfile(name string, profile *Profile) {
	if name == DefaultProfile {
		c.Profile = *profile

		return
	}

	if c.Profiles == nil {
		c.Profiles = map[string]*Profile{}
	}

	c.Profiles[name] = profile
}

func GetEnvProfile() *Profile {
	var profile Profile

	if url := os.Getenv(EnvProfileURL); url != "" {
		profile.ApiUrl = &url
	}

	if apiKey := os.Getenv(EnvProfileAPIKey); apiKey != "" {
		profile.ApiKey = &apiKey
	}

	if cloudProject := os.Getenv(EnvProfileCloudProject); cloudProject != "" {
		i, _ := strconv.Atoi(cloudProject)
		profile.CloudProject = &i
	}

	if cloudRegion := os.Getenv(EnvProfileCloudRegion); cloudRegion != "" {
		i, _ := strconv.Atoi(cloudRegion)
		profile.CloudRegion = &i
	}

	return &profile
}

func MergeProfiles(original *Profile, profiles ...*Profile) *Profile {
	var result = &Profile{
		ApiKey:       original.ApiKey,
		ApiUrl:       original.ApiUrl,
		CloudRegion:  original.CloudRegion,
		CloudProject: original.CloudProject,
	}

	for _, profile := range profiles {
		if profile.ApiKey != nil {
			result.ApiKey = pointer.To(*profile.ApiKey)
		}

		if profile.ApiUrl != nil {
			result.ApiUrl = pointer.To(*profile.ApiUrl)
		}

		if profile.CloudRegion != nil {
			result.CloudRegion = pointer.To(*profile.CloudRegion)
		}

		if profile.CloudProject != nil {
			result.CloudProject = pointer.To(*profile.CloudProject)
		}
	}

	return result
}

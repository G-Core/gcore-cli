package core

import (
	"context"
	"net/http"
	"os"

	"github.com/G-core/gcore-cli/internal/config"
)

const metaKey = iota

// meta contains information about global flags and cli configuration
type meta struct {
	cfg *config.Config
	ctx context.Context

	// Global flags
	flagConfig  string
	flagProfile string
	flagForce   bool
	flagWait    bool

	// Auth function
	authFunc func(ctx context.Context, req *http.Request) error
}

func injectMeta(ctx context.Context, m meta) context.Context {
	return context.WithValue(ctx, metaKey, m)
}

func extractMeta(ctx context.Context) meta {
	return ctx.Value(metaKey).(meta)
}

func ExtractConfig(ctx context.Context) *config.Config {
	return extractMeta(ctx).cfg
}

func ExtractConfigPath(ctx context.Context) (string, error) {
	path := extractMeta(ctx).flagConfig
	if len(path) != 0 {
		return path, nil
	}

	path = os.Getenv(config.EnvConfigPath)
	if len(path) != 0 {
		return path, nil
	}

	return config.GetConfigPath()
}

func ExtractProfile(ctx context.Context) string {
	profileName := extractMeta(ctx).flagProfile
	if len(profileName) > 0 {
		return profileName
	}

	profile := os.Getenv("GCORE_PROFILE")
	if len(profile) > 0 {
		return profile
	}

	cfg := ExtractConfig(ctx)
	if len(cfg.ActiveProfile) > 0 {
		return cfg.ActiveProfile
	}

	return config.DefaultProfile
}

// GetClientProfile returns current profile for client merged from config, envs and flag variables
func GetClientProfile(ctx context.Context) (*config.Profile, error) {
	name := ExtractProfile(ctx)
	cfg := ExtractConfig(ctx)

	profile, err := cfg.GetProfile(name)
	if err != nil {
		return nil, err
	}

	envProfile := config.GetEnvProfile()

	return config.MergeProfiles(profile, envProfile), nil
}

func ExtractAuthFunc(ctx context.Context) func(ctx context.Context, req *http.Request) error {
	return extractMeta(ctx).authFunc
}

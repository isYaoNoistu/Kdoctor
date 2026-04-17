package profile

import "kdoctor/internal/config"

func ApplyBuiltin(cfg config.Config, name string) config.Config {
	profiles := BuiltinProfiles()
	if name == "" {
		name = cfg.DefaultProfile
	}
	profile, ok := profiles[name]
	if !ok {
		return cfg
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]config.ProfileConfig{}
	}
	current := cfg.Profiles[name]
	cfg.Profiles[name] = config.Merge(config.Config{
		Profiles: map[string]config.ProfileConfig{name: profile},
	}, config.Config{
		Profiles: map[string]config.ProfileConfig{name: current},
	}).Profiles[name]
	return cfg
}

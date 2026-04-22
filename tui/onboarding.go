package tui

func Onboarding() (*Config, error) {
	config := &Config{
		Providers: map[string]ProviderConfig{},
	}

	err := ConnectProvider(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func ConnectProvider(config *Config) error {
	PrintHarnessMessage("Input new provider name:")
	name := GetUserInput()

	PrintHarnessMessage("Input provider API URL:")
	url := GetUserInput()

	PrintHarnessMessage("Input provider API key:")
	key := GetUserInput()

	PrintHarnessMessage("Input default model:")
	model := GetUserInput()

	PrintHarnessMessage("Skip SSL verify? (y/n)")
	skipVerify := GetUserInput() == "y"

	provider := ProviderConfig{
		Url:        url,
		Key:        key,
		Model:      model,
		SkipVerify: skipVerify,
	}
	config.Providers[name] = provider
	config.CurrentProvider = name

	return config.Save()
}

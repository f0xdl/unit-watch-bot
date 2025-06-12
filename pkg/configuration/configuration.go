package configuration

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"reflect"
)

// Objects

type IConfig interface {
	Validate() error
	Name() string
	ToString() string
}

// Public

// LoadConfig read configuration and transfer to Config model.
func LoadConfig[T IConfig](path string, config T) error {
	if reflect.ValueOf(config).IsNil() {
		return fmt.Errorf("config cannot be nil")
	}
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var data map[string]interface{}
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		return fmt.Errorf("failed to decode config file: %w", err)
	}

	sectionData, ok := data[config.Name()]
	if !ok {
		return fmt.Errorf("section %s not found in config file", config.Name())
	}

	sectionBytes, err := yaml.Marshal(sectionData)
	if err != nil {
		return fmt.Errorf("failed to marshal section data: %w", err)
	}

	err = yaml.Unmarshal(sectionBytes, config)
	if err != nil {
		return fmt.Errorf("failed to unmarshal section data: %w", err)
	}

	err = config.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate config file: %w", err)
	}
	return nil
}

// LoadIConfig read configuration and transfer to Config model.
func LoadIConfig(path string, config IConfig) error {
	if reflect.ValueOf(config).IsNil() {
		return fmt.Errorf("config cannot be nil")
	}
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var data map[string]interface{}
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		return fmt.Errorf("failed to decode config file: %w", err)
	}

	sectionData, ok := data[config.Name()]
	if !ok {
		return fmt.Errorf("section %s not found in config file", config.Name())
	}

	sectionBytes, err := yaml.Marshal(sectionData)
	if err != nil {
		return fmt.Errorf("failed to marshal section data: %w", err)
	}

	err = yaml.Unmarshal(sectionBytes, config)
	if err != nil {
		return fmt.Errorf("failed to unmarshal section data: %w", err)
	}

	err = config.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate config file: %w", err)
	}
	return nil
}

func LoadConfigs(path string, args []IConfig) error {
	for _, config := range args {
		err := LoadIConfig(path, config)
		if err != nil {
			return err
		}
	}
	return nil
}

func ValidateLoggerLevel(level string) bool {
	switch level {
	case
		"debug",
		"info",
		"warn",
		"error":
		return true
	default:
		return false
	}
}

// Internal

// Plugs

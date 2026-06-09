package config

import (
	"os"
	"reflect"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"
)

func loadIntoConfig(config *Config, configPath string) error {
	if err := applyDefaults(reflect.ValueOf(config).Elem()); err != nil {
		return err
	}
	if configPath != "" {
		if data, err := os.ReadFile(configPath); err == nil {
			if err := yaml.Unmarshal(data, config); err != nil {
				return err
			}
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	if err := applyEnv(reflect.ValueOf(config).Elem()); err != nil {
		return err
	}
	loadOidcProvidersFromEnv(config)
	return nil
}

func applyDefaults(value reflect.Value) error {
	value = indirect(value)
	if !value.IsValid() || value.Kind() != reflect.Struct {
		return nil
	}

	valueType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		structField := valueType.Field(i)
		if !field.CanSet() || structField.Tag.Get("yaml") == "-" {
			continue
		}
		if field.Kind() == reflect.Struct {
			if err := applyDefaults(field); err != nil {
				return err
			}
		}
		defaultValue := structField.Tag.Get("default")
		if defaultValue == "" {
			continue
		}
		if err := setFromString(field, defaultValue); err != nil {
			return err
		}
	}
	return nil
}

func applyEnv(value reflect.Value) error {
	value = indirect(value)
	if !value.IsValid() || value.Kind() != reflect.Struct {
		return nil
	}

	valueType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		structField := valueType.Field(i)
		if !field.CanSet() || structField.Tag.Get("yaml") == "-" {
			continue
		}
		if field.Kind() == reflect.Struct {
			if err := applyEnv(field); err != nil {
				return err
			}
		}
		envName := structField.Tag.Get("env")
		if envName == "" {
			continue
		}
		if envValue, ok := os.LookupEnv(envName); ok {
			if err := setFromString(field, envValue); err != nil {
				return err
			}
		}
	}
	return nil
}

func loadOidcProvidersFromEnv(config *Config) {
	providers := map[int]*oidcProviderConfig{}
	for _, entry := range os.Environ() {
		key, value, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}

		index, fieldPart, ok := oidcEnvParts(key)
		if !ok {
			continue
		}
		if value == "" {
			continue
		}
		provider := providers[index]
		if provider == nil {
			provider = &oidcProviderConfig{}
			providers[index] = provider
		}

		switch fieldPart {
		case "NAME":
			provider.Name = value
		case "DISPLAY_NAME":
			provider.DisplayName = value
		case "CLIENT_ID":
			provider.ClientID = value
		case "CLIENT_SECRET":
			provider.ClientSecret = value
		case "ENDPOINT":
			provider.Endpoint = value
		case "USERNAME_CLAIM":
			provider.UsernameClaim = value
		case "SCOPES":
			provider.Scopes = splitList(value)
		}
	}

	if len(providers) == 0 {
		return
	}
	config.Security.OidcProviders = make([]oidcProviderConfig, 0, len(providers))
	for i := 0; ; i++ {
		provider, ok := providers[i]
		if !ok {
			if i >= len(providers) {
				break
			}
			continue
		}
		config.Security.OidcProviders = append(config.Security.OidcProviders, *provider)
	}
}

func oidcEnvParts(key string) (int, string, bool) {
	if strings.HasPrefix(key, "WAKAPI_OIDC_PROVIDERS_") {
		rest := strings.TrimPrefix(key, "WAKAPI_OIDC_PROVIDERS_")
		indexPart, fieldPart, ok := strings.Cut(rest, "_")
		if !ok {
			return 0, "", false
		}
		index, err := strconv.Atoi(indexPart)
		return index, fieldPart, err == nil
	}

	if strings.HasPrefix(key, "WAKAPI_SECURITY_OIDCPROVIDERS_") {
		rest := strings.TrimPrefix(key, "WAKAPI_SECURITY_OIDCPROVIDERS_")
		indexPart, fieldPart, ok := strings.Cut(rest, "_")
		if !ok {
			return 0, "", false
		}
		index, err := strconv.Atoi(indexPart)
		return index, configorOidcFieldName(fieldPart), err == nil
	}

	return 0, "", false
}

func configorOidcFieldName(field string) string {
	switch field {
	case "DISPLAYNAME":
		return "DISPLAY_NAME"
	case "CLIENTID":
		return "CLIENT_ID"
	case "CLIENTSECRET":
		return "CLIENT_SECRET"
	case "USERNAMECLAIM":
		return "USERNAME_CLAIM"
	default:
		return field
	}
}

func setFromString(field reflect.Value, value string) error {
	field = indirect(field)
	if !field.CanSet() {
		return nil
	}
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Bool:
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(parsed)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(value, 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetInt(parsed)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsed, err := strconv.ParseUint(value, 0, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetUint(parsed)
	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(value, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetFloat(parsed)
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String {
			field.Set(reflect.ValueOf(splitList(value)))
		}
	}
	return nil
}

func splitList(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func indirect(value reflect.Value) reflect.Value {
	for value.IsValid() && value.Kind() == reflect.Pointer && !value.IsNil() {
		value = value.Elem()
	}
	return value
}

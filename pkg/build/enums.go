package build

import (
	"fmt"
	"strings"
)

type ApplicationType uint

const (
	applicationTypeNil ApplicationType = iota
	applicationTypeHelm
	applicationTypeTerraform
)

var ApplicationTypeEnum = NewEnum[ApplicationType](map[ApplicationType]string{
	applicationTypeHelm:      "helm",
	applicationTypeTerraform: "terraform",
})

func (s *ApplicationType) UnmarshalYAML(
	unmarshal func(interface{}) error,
) error {
	return ApplicationTypeEnum.Unmarshal(unmarshal, s)
}

type CloudProviderType uint

const (
	cloudProviderTypeNil CloudProviderType = iota
	cloudProviderTypeGcp
)

var (
	CloudProviderTypeEnum = NewEnum[CloudProviderType](map[CloudProviderType]string{
		cloudProviderTypeGcp: "gcp",
	})
)

func (s *CloudProviderType) UnmarshalYAML(
	unmarshal func(interface{}) error,
) error {
	return CloudProviderTypeEnum.Unmarshal(unmarshal, s)
}

type SecretProviderType uint

const (
	secretProviderTypeNil SecretProviderType = iota
	secretProviderTypeGcp
	secretProviderTypeGithub
)

var (
	SecretProviderTypeEnum = NewEnum[SecretProviderType](map[SecretProviderType]string{
		secretProviderTypeGcp:    "gcp",
		secretProviderTypeGithub: "github-actions",
	})
)

func (s *SecretProviderType) UnmarshalYAML(
	unmarshal func(interface{}) error,
) error {
	return SecretProviderTypeEnum.Unmarshal(unmarshal, s)
}

type Enum[T comparable] struct {
	keyToValue map[T]string
	valueToKey map[string]T
	options    []string
}

func NewEnum[T comparable](keyToValue map[T]string) Enum[T] {
	valueToKey := map[string]T{}
	var options []string
	for k, v := range keyToValue {
		valueToKey[v] = k
		options = append(options, v)
	}
	return Enum[T]{
		keyToValue: keyToValue,
		valueToKey: valueToKey,
		options:    options,
	}
}

func (e Enum[T]) FromString(value string) (T, bool) {
	key, present := e.valueToKey[value]
	return key, present
}

func (e Enum[T]) ToString(key T) (string, bool) {
	value, present := e.keyToValue[key]
	return value, present
}

func (e Enum[T]) InvalidEnumValue(value string) error {
	var values []string
	for _, v := range e.options {
		values = append(values, v)
	}
	return fmt.Errorf("`%s` is not one of (%s)", value, strings.Join(values, ", "))
}

func (e Enum[T]) Unmarshal(
	unmarshal func(interface{}) error,
	result *T,
) error {
	var value string
	err := unmarshal(&value)
	if err != nil {
		return err
	}

	spType, present := e.FromString(value)
	if !present {
		return e.InvalidEnumValue(value)
	}

	*result = spType
	return nil
}

func (e Enum[T]) Marshal(key T) interface{} {
	return e.keyToValue[key]
}

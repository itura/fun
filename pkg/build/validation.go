package build

import (
	"fmt"
	"github.com/itura/fun/pkg/fun"
	"reflect"
	"strings"
)

type Validator interface {
	Validate(key string) ValidationErrors
}

type ValidationErrors struct {
	key      string
	errors   fun.Config[[]error]
	children []ValidationErrors
}

func NewValidationErrors(key string) ValidationErrors {
	return ValidationErrors{
		key:    key,
		errors: fun.NewConfig[[]error](),
	}
}

func (v ValidationErrors) IsPresent() bool {
	if len(v.errors) > 0 {
		return true
	}
	for _, child := range v.children {
		if child.IsPresent() {
			return true
		}
	}
	return false
}

func (v ValidationErrors) Put(key string, err error) ValidationErrors {
	value, present := v.errors[key]
	if !present {
		value = []error{}
	}
	v.errors = v.errors.Set(key, append(value, err))
	return v
}

func (v ValidationErrors) PutChild(child ValidationErrors) ValidationErrors {
	if child.IsPresent() {
		v.children = append(v.children, child)
	}
	return v
}

func (v ValidationErrors) Error() string {
	builder := &strings.Builder{}
	builder.WriteString("--- Validation errors\n")
	builder = v.toString(builder, "")
	return builder.String()
}

func (v ValidationErrors) toString(builder *strings.Builder, indent string) *strings.Builder {
	if !v.IsPresent() {
		return builder
	}
	if v.key != "" {
		builder.WriteString(fmt.Sprintf("%s%s:\n", indent, v.key))
		indent += "  "
	}
	for key, errs := range v.errors {
		builder.WriteString(fmt.Sprintf("%s%s: ", indent, key))
		for i, err := range errs {
			builder.WriteString(err.Error())
			if i < len(errs)-1 {
				builder.WriteString(", ")
			}
		}
		builder.WriteString("\n")
	}
	for _, child := range v.children {
		child.toString(builder, indent)
	}
	return builder
}

func (v ValidationErrors) Validate(parent interface{}) ValidationErrors {
	return v.ValidateTags(parent).ValidateNested(parent)
}

func (v ValidationErrors) ValidateTags(parent interface{}) ValidationErrors {
	type_ := reflect.TypeOf(parent)
	if type_.Kind().String() == "map" {
		return v
	}
	value := reflect.ValueOf(parent)
	for i := 0; i < type_.NumField(); i++ {
		field := type_.Field(i)
		validations, present := field.Tag.Lookup("validate")
		if present {
			yamlKey, present := field.Tag.Lookup("yaml")
			if !present {
				yamlKey = strings.ToLower(field.Name)
			}
			yamlKey = strings.Split(yamlKey, ",")[0]
			for _, validation := range strings.Split(validations, ",") {
				switch validation {
				case "required":
					if value.FieldByName(field.Name).IsZero() {
						v = v.Put(yamlKey, eMissingRequiredField)
					}
				}
			}
		}
	}
	return v
}

func (v ValidationErrors) ValidateNested(parent interface{}) ValidationErrors {
	type_ := reflect.TypeOf(parent)
	if type_.Kind().String() == "map" {
		return v
	}
	value := reflect.ValueOf(parent)
	for i := 0; i < type_.NumField(); i++ {
		field := type_.Field(i)
		// https://stackoverflow.com/questions/41694647/how-do-i-use-reflect-to-check-if-the-type-of-a-struct-field-is-interface
		if field.Type.Implements(reflect.TypeOf((*Validator)(nil)).Elem()) {
			if value.FieldByName(field.Name).IsZero() {
				continue
			}
			yamlKey, present := field.Tag.Lookup("yaml")
			if !present {
				yamlKey = strings.ToLower(field.Name)
			}
			yamlKey = strings.Split(yamlKey, ",")[0]
			result := value.
				FieldByName(field.Name).
				MethodByName("Validate").
				Call([]reflect.Value{
					reflect.ValueOf(yamlKey),
				})
			validationErrors := result[0].Interface()
			v = v.PutChild(validationErrors.(ValidationErrors))
		}
	}
	return v
}

type MissingSecretProvider struct{}

func (m MissingSecretProvider) Error() string {
	return "Invalid secret provider reference"
}

type InvalidCloudProvider struct{ Message string }

func CloudProviderMissingField(t string) error {
	return InvalidCloudProvider{
		fmt.Sprintf("required for cloud provider of type %s", t),
	}
}

func (m InvalidCloudProvider) Error() string {
	return m.Message
}

var (
	eMissingRequiredField = fun.Error("required")
)

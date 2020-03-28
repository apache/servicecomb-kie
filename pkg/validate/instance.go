package validate

const key = "key"

var defaultValidator = NewValidator()

// custom validate rules
// please use different tag names from third party tags
var customRules = []*RegexValidateRule{
	NewRule(key, `^[a-zA-Z0-9]*$|^[a-zA-Z0-9][a-zA-Z0-9_\-.]*[a-zA-Z0-9]$`, &Option{Min: 1, Max: 256}),
	NewRule("valueType", `^(ini|json|text|yaml|properties){0,1}$`, nil),
	NewRule("kvStatus", `^(enabled|disabled){0,1}$`, nil),
}

// tags of third party validate rules we used, for error translation
var thirdPartyTags = []string{
	"min", "max", "length",
}

// Init initializes validate
func Init() error {
	for _, r := range customRules {
		if err := defaultValidator.RegisterRule(r); err != nil {
			return err
		}
	}
	for _, t := range thirdPartyTags {
		if err := defaultValidator.AddErrorTranslation4Tag(t); err != nil {
			return err
		}
	}
	return nil
}

// Validate validates data
func Validate(v interface{}) error {
	return defaultValidator.Validate(v)
}

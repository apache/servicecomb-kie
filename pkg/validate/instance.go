package validate

var defaultValidator = NewValidator()

const (
	key                   = "key"
	commonNameRegexString = `^[a-zA-Z0-9]*$|^[a-zA-Z0-9][a-zA-Z0-9_\-.]*[a-zA-Z0-9]$`
	asciiRegexString      = `^[\x00-\x7F]*$`
)

// custom validate rules
// please use different tag names from third party tags
var customRules = []*RegexValidateRule{
	NewRule(key, commonNameRegexString, &Option{Min: 1, Max: 128}),
	NewRule("getKey", commonNameRegexString, &Option{Max: 128}),
	NewRule("commonName", commonNameRegexString, &Option{Min: 1, Max: 256}),
	NewRule("valueType", `^$|^(ini|json|text|yaml|properties)$`, nil),
	NewRule("kvStatus", `^$|^(enabled|disabled)$`, nil),
	NewRule("value", asciiRegexString, &Option{Max: 2097152}), //ASCII, 2M
	NewRule("labelKV", commonNameRegexString, &Option{Max: 32}),
	NewRule("check", asciiRegexString, &Option{Max: 1048576}), //ASCII, 1M
}

// tags of third party validate rules we used, for error translation
var thirdPartyTags = []string{
	"min", "max", "length", "uuid",
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

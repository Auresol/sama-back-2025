package utils

import (
	"log"
	"regexp"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

// Validate is the global validator instance.
// It's exported so other packages can use it.
var (
	Validate         *validator.Validate
	trans            ut.Translator
	classroomPattern *regexp.Regexp
)

// init function runs automatically when this package is imported.
// This is where you configure your global validator.
func init() {
	Validate = validator.New()

	// Create a new English translator
	english := en.New()
	uni := ut.New(english, english)    // You can add more locales if needed
	trans, _ = uni.GetTranslator("en") // Get the English translator

	en_translations.RegisterDefaultTranslations(Validate, trans)
	// Register all your custom validation functions here.
	// For your classroom example:
	if err := Validate.RegisterValidation("classroomregex", validateClassroomRegex); err != nil {
		log.Fatalf("Failed to register 'classroomregex' validator: %v", err)
	}

	Validate.RegisterTranslation("classroomregex", trans, func(ut ut.Translator) error {
		return ut.Add("classroomregex", "{0} must be in the format 'X/Y' where X and Y are positive integer less than 100 (1-99)", false)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(fe.Tag(), fe.Field()) // {0} will be replaced by fe.Field()
		return t
	})

	// Register other custom validators as needed, for example:
	// if err := Validate.RegisterValidation("customEmail", validateCustomEmail); err != nil {
	//     log.Fatalf("Failed to register 'customEmail' validator: %v", err)
	// }

	// You can also register custom tag name functions, struct level validations, etc.
	// Example: Use JSON tag names in error messages
	// Validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
	// 	name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	// 	if name == "-" {
	// 		return ""
	// 	}
	// 	return name
	// })

	classroomPattern = regexp.MustCompile(`^(?:[1-9]|[1-9][0-9])\/(?:[1-9]|[1-9][0-9])$`)
}

// validateClassroomRegex is the custom validation logic for the classroom format.
func validateClassroomRegex(fl validator.FieldLevel) bool {
	// Compile the regex once. For simplicity, we're doing it here.
	// For very complex or frequently used regexes, you might compile them
	// as package-level variables within this package.
	return classroomPattern.MatchString(fl.Field().String())
}

// Add other custom validation functions here if needed
// func validateCustomEmail(fl validator.FieldLevel) bool {
//     // ... custom email validation logic ...
// }

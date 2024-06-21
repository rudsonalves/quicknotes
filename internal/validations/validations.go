package validations

type FormValidator struct {
	FieldErrors map[string]string
}

func (fv *FormValidator) AddFieldError(field, message string) {
	if fv.FieldErrors == nil {
		fv.FieldErrors = make(map[string]string)
	}
	fv.FieldErrors[field] = message
}

func (fv *FormValidator) Valid() bool {
	return len(fv.FieldErrors) == 0
}

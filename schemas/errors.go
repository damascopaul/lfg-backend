package schemas

type BodyError struct {
	Message     string       `json:"message,omitempty"`
	FieldErrors []FieldError `json:"field_errors,omitempty"`
}

type FieldError struct {
	Name  string
	Error string
}

type ValidationError struct {
	Message string
	Errors  []FieldError
}

func (e *ValidationError) Error() string {
	return e.Message
}

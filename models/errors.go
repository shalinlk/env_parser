package models

type EnvError struct {
	msg   string
	field string
}

func (e EnvError) Error() string {
	return e.msg + "; field : " + e.field
}

func (e *EnvError) Field(field string) *EnvError {
	e.field = field
	return e
}

func MandatoryValueMissing() *EnvError {
	e := EnvError{}
	e.msg = "mandatory value missing in environment"
	return &e
}

func OptionalValueMissing() *EnvError {
	e := EnvError{}
	e.msg = "optional value missing in environment"
	return &e
}

func InvalidTag() *EnvError {
	e := EnvError{}
	e.msg = "invalid tag"
	return &e
}

func InvalidValue() *EnvError {
	e := EnvError{}
	e.msg = "invalid value from environment / default"
	return &e
}

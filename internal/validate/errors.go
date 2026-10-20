package validate

type Errors map[string]string

func (e Errors) Check(valid bool, field, message string) {
	if _, failed := e[field]; valid || failed {
		return
	}

	e[field] = message
}

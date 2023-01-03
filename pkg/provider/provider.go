package provider

type Provider interface {
	Install() (Output, error)
	Uninstall() error
}

type Output map[string]interface{}

func generateOutputs(input interface{}, templates map[string]string) (Output, error) {
	result := Output{}
	for key, template := range templates {
		value, err := generateOutput(input, template)
		if err != nil {
			return nil, err
		}
		result[key] = value
	}
	return nil, nil
}

func generateOutput(input interface{}, template string) (string, error) {
	return "", nil
}

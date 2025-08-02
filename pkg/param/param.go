package param

func ReadDefault(name string) (string, error) {
	content, err := Defaults.ReadFile(name)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

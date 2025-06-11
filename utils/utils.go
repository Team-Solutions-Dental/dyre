package utils

import "errors"

func Get_Map_Keys(m map[string]any) []string {
	count := len(m)
	keys := make([]string, count)
	i := 0
	for key := range m {
		keys[i] = key
		i += 1
	}
	return keys
}

func Array_Contains(a []string, l string) bool {
	for _, v := range a {
		if l == v {
			return true
		}
	}
	return false
}

func WrapErrorArray(errs []string) []error {
	var output []error
	for _, e := range errs {
		output = append(output, errors.New(e))
	}
	return output
}

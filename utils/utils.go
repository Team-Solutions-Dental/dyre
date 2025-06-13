package utils

import (
	"errors"
)

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

// func deepEqualStringArray(arr1 []string, arr2 []string) []error {
// 	errors := []error{}
// 	for _, k := range arr1 {
// 		if !Array_Contains(arr2, k) {
// 			errors = append(errors, fmt.Errorf("%s not found in %v\n", k, arr2))
// 		}
// 	}
// 	for _, k := range arr2 {
// 		if !Array_Contains(arr1, k) {
// 			errors = append(errors, fmt.Errorf("%s not found in %v\n", k, arr1))
// 		}
// 	}
// 	return errors
// }

// func openJSON(path string) []byte {
// 	jsonFile, err := os.Open(path)
// 	if err != nil {
// 		log.Panic(err)
// 		return nil
// 	}
// 	defer jsonFile.Close()
//
// 	byteValue, err := io.ReadAll(jsonFile)
// 	if err != nil {
// 		log.Panic(err)
// 		return nil
// 	}
//
// 	return byteValue
//
// }

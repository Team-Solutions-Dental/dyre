package transpiler

import (
	"errors"
	"fmt"
)

var builtins = map[string]func(ir *IR, args ...string) (string, error){
	"len": func(ir *IR, args ...string) (string, error) {
		if len(args) != 1 {
			return "", errors.New(fmt.Sprintf("wrong number of arguments. got=%d, want=1",
				len(args)))
		}

		return fmt.Sprintf("LEN(%s)", args[0]), nil
	},
	"date": func(ir *IR, args ...string) (string, error) {
		if len(args) != 1 {
			return "", errors.New(fmt.Sprintf("wrong number of arguments. got=%d, want=1",
				len(args)))
		}

		return fmt.Sprintf("CONVERT(date, %s, 23)", args[0]), nil
	},
}

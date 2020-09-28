package main

import (
	"os"

	"gopkg.in/yaml.v2"
)

func main() {
	_ = yaml.NewEncoder(os.Stdout).
		Encode(map[string]interface{}{
			"foo": "bar",
			"foz": "baz",
		})
}

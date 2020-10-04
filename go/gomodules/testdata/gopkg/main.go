package main

import (
	"os"

	"gopkg.in/yaml.v1"
)

func main() {
	b, err := yaml.Marshal(map[string]interface{}{
		"foo": "bar",
		"foz": "baz",
	})
	if err != nil {
		panic(err)
	}
	_, _ = os.Stdout.Write(b)
}

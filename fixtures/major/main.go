package main

import (
	"fmt"

	"github.com/caarlos0/env/v6"
)

type cfg struct {
	Name string `env:"NAME" envDefault:"World"`
}

func main() {
	var c cfg
	if err := env.Parse(&c); err != nil {
		panic(err)
	}
	fmt.Printf("hello %q\n", c.Name)
}

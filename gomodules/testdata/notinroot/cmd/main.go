package main

import (
	"fmt"

	"github.com/pkg/errors"
)

func main() {
	err := errors.New("kaboom")
	fmt.Println(err)
}

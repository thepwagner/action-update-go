package main

import (
	"fmt"

	_ "golang.org/x/crypto/hkdf"
)

func main() {
	fmt.Println("hkdf imported")
}

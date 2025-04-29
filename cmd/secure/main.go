package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "password argument is missing")
		os.Exit(1)
	}
	h, err := bcrypt.GenerateFromPassword([]byte(os.Args[1]), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintln(os.Stderr, "couldn't encrypt password")
		os.Exit(1)
	}

	fmt.Fprint(os.Stdout, string(h))
}

package internal

import (
	"fmt"
	"os"
)

func Error(message string, err error) {
	fmt.Println(message)
	fmt.Println(err)
	os.Exit(1)
}

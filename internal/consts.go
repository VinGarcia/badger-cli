package internal

import "fmt"

var ErrUnrecognizedCmd = fmt.Errorf("unrecognized command")

var ErrRecordNotFound = fmt.Errorf("record not found")

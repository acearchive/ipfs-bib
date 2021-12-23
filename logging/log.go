package logging

import (
	"log"
	"os"
)

var (
	Error   = log.New(os.Stderr, "", 0)
	Verbose = log.New(os.Stderr, "", 0)
)

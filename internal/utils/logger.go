package utils

import (
	"log"
	"os"
)

var Logger = log.New(os.Stdout, "[BOT] ", log.LstdFlags|log.Lshortfile)

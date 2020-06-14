package main

import (
	"github.com/pmurley/mida/log"
)

func main() {
	initViperConfig()
	log.InitGlobalLogger("mida.log")

	rootCmd := getRootCommand()
	err := rootCmd.Execute()
	if err != nil {
		log.Error(err)
	}

	return
}

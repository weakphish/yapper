package main

import (
	"github.com/weakphish/yapper/cmd"
	"github.com/weakphish/yapper/internal/config"
)

func main() {
	config.InitConfig()
	cmd.Execute()
}

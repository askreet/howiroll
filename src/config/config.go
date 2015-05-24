package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	ASGName string
	DryRun  bool
}

func ParseConfig() *Config {
	config := &Config{}

	flag.StringVar(&config.ASGName, "asg-name", "", "")
	flag.BoolVar(&config.DryRun, "dry-run", false, "Do all sanity checks and list instances that would be replaced, then quit.")

	flag.Parse()

	var missingArgs []string
	missingArgs = make([]string, 0)

	if config.ASGName == "" {
		missingArgs = append(missingArgs, "-asg-name")
	}

	if len(missingArgs) != 0 {
		fmt.Println("Missing required arguments:", strings.Join(missingArgs, ", "))
		os.Exit(1)
	}

	return config
}

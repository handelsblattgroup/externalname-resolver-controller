package options

import (
	"fmt"
	"os"
	"time"
)

var Current = NewOptions()

const defaultResync = time.Minute

func NewOptions() *Options {
	options := new(Options)

	options.ResyncInterval = defaultResync

	defaultConfigPath := os.Getenv("KUBECONFIG")

	if defaultConfigPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic("was not able to find the home directory to laod the kubeconfig")
		}
		defaultConfigPath = fmt.Sprintf("%s/.kube/config", homeDir)
	}

	options.Kubeconfig = defaultConfigPath

	return options
}

type Options struct {
	ResyncInterval time.Duration
	Kubeconfig     string
	Debug          bool
	IgnoreLabels   []string
	ClusterDnsIP   string
}

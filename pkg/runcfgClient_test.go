package runcfg_test

import (
	"fmt"
	runcfg "github.com/runcfg/runcfg-go/pkg"
	"sync"
	"testing"
)

type ExampleConfig struct {
	Version string `json:"version"`
	Target  string `json:"target"`
	Enabled string `json:"enabled"`
}

func TestCreate_AndGetConfig(t *testing.T) {
	var config ExampleConfig

	// create runcfg client
	client, err := runcfg.Create()
	if err != nil {
		t.Error(err)
	}
	// fetch remote config and deserialize into ExampleConfig type
	err = client.LoadConfigAsType("1.0.0", &config)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("[.runcfg] loaded remote config")
	fmt.Printf("[.runcfg] Value `version: %s`\n", config.Version)
	fmt.Printf("[.runcfg] Value `target: %s`\n", config.Target)
	fmt.Printf("[.runcfg] Value `enabled: %s`\n", config.Enabled)
}

func TestCreate_AndGetConfig_UsingGoroutine(t *testing.T) {

	// create runcfg client
	client, err := runcfg.Create()
	if err != nil {
		t.Error(err)
	}

	// create channels for goroutine
	configChan := make(chan ExampleConfig, 1)
	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		var config ExampleConfig

		// fetch remote config on interval
		err := client.LoadConfigAsType("1.0.0", &config)
		if err != nil {
			errChan <- err
		}
		configChan <- config

		defer wg.Done()
		close(configChan)
		close(errChan)
	}()
	wg.Wait()

	// get config from channel
	config := <-configChan

	fmt.Println("[.runcfg] loaded remote config")
	fmt.Printf("[.runcfg] Value `version: %s`\n", config.Version)
	fmt.Printf("[.runcfg] Value `target: %s`\n", config.Target)
	fmt.Printf("[.runcfg] Value `enabled: %s`\n", config.Enabled)
}

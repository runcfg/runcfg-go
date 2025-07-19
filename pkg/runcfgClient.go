package runcfg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	TargetApi = "https://runcfg.com"
)

type Client struct {
	ProjectId      string `json:"projectId"`
	ClientToken    string `json:"clientToken"`
	previousUpdate string
}

type ConfigUpdate struct {
	Enabled     bool   `json:"enabled"`
	LastUpdated string `json:"lastUpdate"`
}

func Create(name string, watch bool) (Client, error) {
	var file []byte
	var err error

	if name != "" {
		pwd, _ := os.Getwd()
		file, err = os.ReadFile(pwd + "/" + name + ".runcfg")
		if err != nil {
			return Client{ProjectId: "", ClientToken: ""},
				errors.New("[.runcfg] Failed to load local .runcfg file at path (" + pwd + "/" + name + ".runcfg)")
		}
	} else {
		file, err = os.ReadFile(".runcfg")
		if err != nil {
			return Client{ProjectId: "", ClientToken: ""}, errors.New("[.runcfg] Failed to load local .runcfg")
		}
	}

	var clientConfig Client
	err = json.Unmarshal(file, &clientConfig)
	fmt.Println(err)
	if err != nil {
		return Client{ProjectId: "", ClientToken: ""},
			errors.New("[.runcfg] Failed to load remote config")
	}
	client := Client{
		ProjectId:   clientConfig.ProjectId,
		ClientToken: clientConfig.ClientToken,
	}

	// if watch specified and after first successful config fetch we start the update timer
	// to watch every 15 seconds
	if watch {
		fmt.Println("[.runcfg] Watching for changes every 15 seconds...")
		ticker := time.NewTicker(15 * time.Second)
		quit := make(chan struct{})
		go func() {
			for {
				select {
				case <-ticker.C:
					client.Updated("1.0.0")
				case <-quit:
					ticker.Stop()
					return
				}
			}
		}()
	}

	return client, nil
}

func (c *Client) LoadConfigAsType(version string, configType interface{}) error {
	if c.ProjectId == "" || c.ClientToken == "" {
		return errors.New("[.runcfg] .runcfg values are not valid")
	}

	target := TargetApi + "/app/project/" + c.ProjectId + "/view"
	client := &http.Client{}
	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		return errors.New("[.runcfg] Failed to construct request for runcfg")
	}
	req.Header.Set("Authorization", c.ClientToken)
	req.Header.Set("Version", version)
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("[.runcfg] Failure requesting config")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New("[.runcfg] Failed to parse response in GetConfig")
	}
	// unescape received string
	unquoted, err := strconv.Unquote(string(body))
	if err != nil {
		return errors.New("[.runcfg] Failed to unescape response config")
	}

	// unmarshal config into ExampleConfig type
	errb := json.Unmarshal([]byte(fmt.Sprintf("%s", unquoted)), configType)
	if errb != nil {
		return errors.New("[.runcfg] Failed to unmarshal config JSON")
	}

	return nil
}

func (c *Client) Updated(version string) (bool, error) {
	if c.ProjectId == "" || c.ClientToken == "" {
		return false, errors.New("[.runcfg] .runcfg values are not valid")
	}

	target := TargetApi + "/app/project/" + c.ProjectId + "/updated"
	client := &http.Client{}
	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		return false, errors.New("[.runcfg] Failed to construct request for runcfg")
	}
	req.Header.Set("Authorization", c.ClientToken)
	req.Header.Set("Version", version)
	resp, err := client.Do(req)
	if err != nil {
		return false, errors.New("[.runcfg] Failure requesting config")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, errors.New("[.runcfg] Failed to parse response in GetConfig")
	}

	// unescape received string
	unquoted, err := strconv.Unquote(string(body))
	if err != nil {
		return false, errors.New("[.runcfg] Failed to unescape response config")
	}

	var configUpdateType ConfigUpdate

	// unmarshal config into ExampleConfig type
	errb := json.Unmarshal([]byte(fmt.Sprintf("%s", unquoted)), configUpdateType)
	if errb != nil {
		return false, errors.New("[.runcfg] Failed to unmarshal config JSON")
	}

	if !configUpdateType.Enabled {
		return false, errors.New("[.runcfg] Config is disabled")
	} else {
		updated, _ := time.Parse(time.RFC3339, configUpdateType.LastUpdated)
		now := time.Now()
		start, _ := time.Parse(time.RFC3339, c.previousUpdate)
		if inTimeSpan(start, now, updated) {
			return true, nil
		} else {
			return false, nil
		}
	}
}

func inTimeSpan(start, end, check time.Time) bool {
	return check.After(start) && check.Before(end)
}

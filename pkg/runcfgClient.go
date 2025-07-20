package runcfg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	TargetApi = "https://runcfg.com"
)

type Client struct {
	ProjectId      string `json:"projectId"`
	ClientToken    string `json:"clientToken"`
	ReqClient      *http.Client
	previousUpdate string
	quit           chan struct{}
	ticker         *time.Ticker
	configType     interface{}
	wg             sync.WaitGroup
}

type ConfigUpdate struct {
	Enabled     bool   `json:"enabled"`
	LastUpdated string `json:"updated"`
}

func Create(name string) (*Client, error) {
	var file []byte
	var err error

	if name != "" {
		pwd, _ := os.Getwd()
		file, err = os.ReadFile(pwd + "/" + name + ".runcfg")
		if err != nil {
			return &Client{ProjectId: "", ClientToken: ""},
				errors.New("[.runcfg] Failed to load local .runcfg file at path (" + pwd + "/" + name + ".runcfg)")
		}
	} else {
		file, err = os.ReadFile(".runcfg")
		if err != nil {
			return &Client{ProjectId: "", ClientToken: ""}, errors.New("[.runcfg] Failed to load local .runcfg")
		}
	}

	var clientConfig Client
	err = json.Unmarshal(file, &clientConfig)
	if err != nil {
		return &Client{ProjectId: "", ClientToken: ""},
			errors.New("[.runcfg] Failed to load remote config")
	}
	client := Client{
		ProjectId:   clientConfig.ProjectId,
		ClientToken: clientConfig.ClientToken,
	}

	// if watch specified and after first successful config fetch we start the update timer
	// to watch every 15 seconds

	client.quit = make(chan struct{})
	client.wg = sync.WaitGroup{}
	client.GetLatestVersion()

	return &client, nil
}

func (c *Client) Watch(seconds time.Duration) {
	if seconds*time.Second < time.Second*5 {
		fmt.Println("[.runcfg] minimum watch interval is 5")
		return
	}

	fmt.Printf("[.runcfg] Watching for changes every %s\n", seconds*time.Second)
	c.ticker = time.NewTicker(seconds * time.Second)
	c.wg.Add(1)
	go func() {
		for {
			select {
			case <-c.ticker.C:
				updated, err := c.Updated("1.0.0")
				if err != nil {
					log.Fatalf("%s", err.Error())
				}
				if updated {
					err := c.LoadConfigAsType("1.0.0", nil)
					if err != nil {
						log.Fatalf("%s", err.Error())
					} else {
						fmt.Println("[.runcfg] Config Updated " + time.Now().String())
					}
				}
			case <-c.quit:
				c.ticker.Stop()
				defer c.wg.Done()
				return
			}
		}
	}()
	c.wg.Wait()
}

func (c *Client) LoadConfigAsType(version string, configType interface{}) error {
	var confType interface{}

	if c.configType == nil && configType == nil {
		return errors.New("[.runcfg] config type not specified")
	} else if c.configType != nil && configType == nil {
		confType = c.configType
	} else if configType != nil {
		confType = configType
		c.configType = configType
	}

	if c.ProjectId == "" || c.ClientToken == "" {
		return errors.New("[.runcfg] .runcfg values are not valid")
	}

	target := TargetApi + "/app/project/" + c.ProjectId + "/view"
	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		return errors.New("[.runcfg] Failed to construct request for runcfg")
	}

	req.Header.Set("Authorization", c.ClientToken)
	req.Header.Add("Version", version)

	resp, err := c.ReqClient.Do(req)
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
		return errors.New("[.runcfg] Failed to unescape response config\n[.runcfg] Response Code: " + strconv.Itoa(resp.StatusCode))
	}

	errb := json.Unmarshal([]byte(fmt.Sprintf("%s", unquoted)), confType)
	if errb != nil {
		return errors.New("[.runcfg] Failed to unmarshal config JSON")
	}

	return nil
}

func (c *Client) GetLatestVersion() error {
	c.ReqClient = &http.Client{}
	latest, err := c.fetchLatest()
	if err != nil {
		return errors.New("[.runcfg] (GetLatest) Failure during fetch " + err.Error())
	}
	if latest.Enabled {
		c.previousUpdate = latest.LastUpdated
	}

	return nil
}

func (c *Client) Updated(version string) (bool, error) {
	latest, err := c.fetchLatest()
	if err != nil {
		return false, errors.New("[.runcfg] Failed to fetch latest: " + err.Error())
	}
	if !latest.Enabled {
		return false, errors.New("[.runcfg] Config is disabled")
	} else {
		updated, _ := time.Parse(time.RFC3339, latest.LastUpdated)
		start, _ := time.Parse(time.RFC3339, c.previousUpdate)
		c.previousUpdate = latest.LastUpdated
		if updated.After(start) {
			return true, nil
		} else {
			return false, nil
		}
	}
}

func (c *Client) fetchLatest() (*ConfigUpdate, error) {
	if c.ProjectId == "" || c.ClientToken == "" {
		return nil, errors.New("[.runcfg] .runcfg values are not valid")
	}

	target := TargetApi + "/app/project/" + c.ProjectId + "/updated"

	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		return nil, errors.New("[.runcfg] Failed to construct request for runcfg")
	}

	req.Header.Add("Authorization", c.ClientToken)
	req.Header.Add("Version", "rcgo-1.0.0")

	resp, err := c.ReqClient.Do(req)
	if err != nil {
		return nil, errors.New("[.runcfg] Failure requesting config; " + err.Error())
	}

	if strconv.Itoa(resp.StatusCode) != "200" {
		return nil, errors.New("[.runcfg] Failed to get updated values for project\n[.runcfg] Response StatusCode: " +
			strconv.Itoa(resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("[.runcfg] Failed to parse response in GetConfig")
	}

	var configUpdateType *ConfigUpdate

	errb := json.Unmarshal([]byte(fmt.Sprintf("%s", body)), &configUpdateType)
	if errb != nil {
		return nil, errors.New("[.runcfg] Failed to unmarshal config JSON\n[.runcfg]" + errb.Error())
	}
	return configUpdateType, nil
}

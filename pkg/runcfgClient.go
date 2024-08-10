package runcfg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

const (
	TargetApi = "https://runcfg.com"
)

type Client struct {
	ProjectId   string `json:"projectId"`
	ClientToken string `json:"clientToken"`
}

func Create() (Client, error) {
	file, err := os.ReadFile(".runcfg")
	var clientConfig Client
	err = json.Unmarshal(file, &clientConfig)
	fmt.Println(err)
	if err != nil {
		return Client{ProjectId: "", ClientToken: ""},
			errors.New("[.runcfg] Failed to load remote config")
	}
	return Client{
		ProjectId:   clientConfig.ProjectId,
		ClientToken: clientConfig.ClientToken,
	}, nil
}

func (c *Client) LoadConfigAsType(version string, configType interface{}) error {
	if c.ProjectId == "" || c.ClientToken == "" {
		return errors.New("[.runcfg] .runcfg values are not valid")
	}

	target := TargetApi + "/app/project/" + c.ProjectId + "/view"
	client := &http.Client{}
	req, err := http.NewRequest("GET", target, nil)
	req.Header.Set("Authorization", c.ClientToken)
	req.Header.Set("Version", version)
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("failure requesting config")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New("failed to parse response in GetConfig")
	}
	// unescape received string
	unquoted, err := strconv.Unquote(string(body))
	if err != nil {
		return errors.New("failed to unescape response config")
	}

	// unmarshal config into ExampleConfig type
	errb := json.Unmarshal([]byte(fmt.Sprintf("%s", unquoted)), configType)
	if errb != nil {
		return errors.New("[.runcfg] Failed to unmarshal config JSON")
	}

	return nil
}

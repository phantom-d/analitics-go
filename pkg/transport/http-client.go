package transport

import (
	"analitics/pkg/config"
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
	"io/ioutil"
	"net/http"
	"time"
)

type HttpClient struct {
	Host      string `mapstructure:"host"`
	User      string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	UserAgent string `mapstructure:"user-agent"`
	conn      *http.Client
}

func (c *HttpClient) Init(cfg map[string]interface{}) TransportInterface {
	err := mapstructure.Decode(cfg, &c)
	if err != nil {
		config.Log().Fatal().Err(err).Msg("Error initialisation HTTP HttpClient!")
	}
	c.conn = &http.Client{
		Timeout: time.Second * 10,
	}
	return c
}

func (c *HttpClient) GetEntities(queue string) (result *Package, err error) {
	p := new(Package)
	uri := "entities/" + queue
	body, err := c.SendRequest(uri, "GET", nil)
	if err != nil {
		config.Log().Error().Err(err).Msg("")
		return
	}
	if string(body) != "[]" {
		err = json.Unmarshal(body, &p)
		if err != nil {
			config.Log().Error().Err(err).Msg("")
		}
		if p.Message != nil {
			err = errors.New(*p.Message)
			config.Log().Error().Err(err).Msg("")
		}
	}
	result = p
	return
}

func (c *HttpClient) ConfirmPackage(queue string, packageID int64) {
	uri := "confirm-package"
	formData := Confirm{
		PackageID: packageID,
		Type:      queue,
	}
	data, err := json.Marshal(formData)
	if err != nil {
		config.Log().Error().Err(err).Msg("")
	}
	_, err = c.SendRequest(uri, "POST", data)
	if err != nil {
		config.Log().Error().Err(err).Msg("")
	}
}

func (c *HttpClient) ResendErrorItems(entityType string, entityIDs []string) bool {
	if entityIDs == nil {
		return false
	}
	uri := "entities/add-to-queue"

	errorItemsData := ErrorItemsData{
		Queue: entityType,
		Guids: entityIDs,
	}
	sign, err := json.Marshal(errorItemsData)
	if err != nil {
		config.Log().Error().Err(err).Msg("")
		return false
	}

	formData := ErrorItems{}
	formData.Header.Source = "analitics"
	formData.Header.SourceDetail = "Import from: " + entityType
	formData.Header.Date = time.Now().Format(time.RFC3339)
	formData.Header.Sign = fmt.Sprintf("%x", md5.Sum(sign))
	formData.Data = errorItemsData

	data, err := json.Marshal(formData)
	if err != nil {
		config.Log().Error().Err(err).Msg("")
		return false
	}
	_, err = c.SendRequest(uri, "POST", data)
	if err != nil {
		config.Log().Error().Err(err).Msg("")
		return false
	}
	return true
}

func (c *HttpClient) SendRequest(uri string, method string, formData []byte) ([]byte, error) {
	url := c.Host + "/" + uri + "/"
	if method == "" {
		method = "GET"
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(formData))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", c.UserAgent)

	if c.User != "" && c.Password != "" {
		req.SetBasicAuth(c.User, c.Password)
	}

	resp, err := c.conn.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		if config.App().Debug {
			config.Log().Debug().
				Dict("data", zerolog.Dict().
					Str("request", spew.Sdump(req)).
					Str("body", string(body)),
				).
				Msg("Response body")
		}
		return nil, err
	}
	return body, nil
}

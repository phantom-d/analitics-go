package transport

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type Package struct {
	PackageID int64    `json:"package_id"`
	Data      []string `json:"data"`
}

type ErrorItems struct {
	Header struct {
		Source       string `json:"source"`
		SourceDetail string `json:"source_detail"`
		Date         string `json:"date"`
		Sign         string `json:"sign"`
	} `json:"header"`
	Data ErrorItemsData `json:"data"`
}

type ErrorItemsData struct {
	Queue string   `json:"queue"`
	Guids []string `json:"guids"`
}

type Confirm struct {
	PackageID int64  `json:"package_id"`
	Type      string `json:"type"`
}

type Client struct {
	Host     string
	User     string
	Password string
	conn     *http.Client
}

func NewClient(cfg map[string]interface{}) *Client {
	c := &Client{}
	if host, ok := cfg["Host"].(string); ok {
		c.Host = host
	} else {
		Logger().Error().Msg("Error: Not defined host for http client!")
		return nil
	}
	if user, ok := cfg["User"].(string); ok {
		c.User = user
		if password, ok := cfg["Pass"].(string); ok {
			c.Password = password
		}
	}

	c.conn = &http.Client{
		Timeout: time.Second * 10,
	}

	return c
}

func (c *Client) GetEntities(queue string) *Package {
	uri := "entities/" + queue
	resp, err := c.SendRequest(uri, "GET", nil)
	if err != nil {
		Logger().Error().Err(err).Msg("")
		return nil
	}
	return resp
}

func (c *Client) ConfirmPackage(queue string, packageID int64) {
	uri := "confirm-package" + queue
	formData := Confirm{
		PackageID: packageID,
		Type:      queue,
	}
	data, err := json.Marshal(formData)
	if err != nil {
		Logger().Error().Err(err).Msg("")
	}
	_, err = c.SendRequest(uri, "POST", data)
	if err != nil {
		Logger().Error().Err(err).Msg("")
	}
}

func (c *Client) ResendErrorItems(entityType string, entityIDs []string) bool {
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
		Logger().Error().Err(err).Msg("")
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
		Logger().Error().Err(err).Msg("")
		return false
	}
	_, err = c.SendRequest(uri, "POST", data)
	if err != nil {
		Logger().Error().Err(err).Msg("")
		return false
	}
	return true
}

func (c *Client) SendRequest(uri string, method string, formData []byte) (*Package, error) {
	p := &Package{}
	url := c.Host + "/" + uri + "/"
	if method == "" {
		method = "GET"
	}
	content, err := ioutil.ReadFile("test.json")
	if err != nil {
		Logger().Error().Err(err).Msg("")
		return nil, err
	}
	content = []byte(os.ExpandEnv(string(content)))
	if err := json.Unmarshal(content, &p); err != nil {
		Logger().Error().Err(err).Msg("")
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(formData))
	if err != nil {
		Logger().Error().Err(err).Msg("")
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "Analitics Exchange")

	if c.User != "" && c.Password != "" {
		req.SetBasicAuth(c.User, c.Password)
	}

	resp, err := c.conn.Do(req)
	if err != nil {
		Logger().Error().Err(err).Msg("")
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Logger().Error().Err(err).Msg("")
		return nil, err
	}

	err = json.Unmarshal(body, &p)
	if err != nil {
		Logger().Error().Err(err).Msg("")
		return nil, err
	}
	return p, nil
}

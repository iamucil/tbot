package tbot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type responseParameters struct {
	MigrateToChatID int `json:"migrate_to_chat_id"`
	RetryAfter      int `json:"retry_after"`
}

type apiResponse struct {
	OK          bool                `json:"ok"`
	Result      json.RawMessage     `json:"result"`
	Description string              `json:"description"`
	ErrorCode   int                 `json:"error_code"`
	Parameter   *responseParameters `json:"parameter,omitempty"`
}

var netTransport = &http.Transport{
	TLSHandshakeTimeout:   10 * time.Second,
	MaxIdleConns:          10,
	IdleConnTimeout:       30 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

func (c *Client) sendRequest(method string, request url.Values, response interface{}) error {
	var err error
	var req *http.Request
	var resp *http.Response
	endPoint := fmt.Sprintf(c.url, method)
	if request == nil {
		req, err = http.NewRequest(http.MethodPost, endPoint, nil)
	} else {
		req, err = http.NewRequest(http.MethodPost, endPoint, strings.NewReader(request.Encode()))

	}
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")

	netClient := &http.Client{
		Timeout:   120 * time.Second,
		Transport: netTransport,
	}

	resp, err = netClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// resp.StatusCode is between 200 and 300.
	// This is because an HTTP status code with the form 2XX signifies a successful HTTP POST request
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var apiResp apiResponse
		err = json.NewDecoder(resp.Body).Decode(&apiResp)
		if err != nil {
			return fmt.Errorf("unable to decode response: %v", err)
		}
		err = resp.Body.Close()
		if err != nil {
			return fmt.Errorf("unable to close response body: %v", err)
		}

		if !apiResp.OK {
			return fmt.Errorf(apiResp.Description)
		}

		return json.Unmarshal(apiResp.Result, response)
	}

	return fmt.Errorf(resp.Status)
}

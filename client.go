package moipayway

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	TestBase = "https://dev.moipayway.co"
	LiveBase = "https://api.moipayway.co"
)

type Client struct {
	APIKey      string
	Environment string
	HTTP        *http.Client
}

type APIError struct {
	Message    string
	HTTPStatus int
	Payload    map[string]any
}

func (e *APIError) Error() string {
	return e.Message
}

func New(apiKey, environment string) (*Client, error) {
	env := strings.ToLower(strings.TrimSpace(environment))
	if env == "" {
		env = "test"
	}
	if env != "test" && env != "live" {
		return nil, fmt.Errorf("environment must be test or live")
	}
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("apiKey is required")
	}
	return &Client{
		APIKey:      strings.TrimSpace(apiKey),
		Environment: env,
		HTTP:        &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (c *Client) BaseURL() string {
	if c.Environment == "live" {
		return LiveBase
	}
	return TestBase
}

func (c *Client) Request(method, path string, body map[string]any, auth bool) (map[string]any, error) {
	url := strings.TrimRight(c.BaseURL(), "/") + "/" + strings.TrimLeft(path, "/")
	var reader io.Reader
	if !strings.EqualFold(method, http.MethodGet) && !strings.EqualFold(method, http.MethodHead) {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(strings.ToUpper(method), url, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if auth {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var decoded map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &decoded); err != nil {
			return nil, &APIError{Message: "Invalid JSON response", HTTPStatus: res.StatusCode}
		}
	}
	if decoded == nil {
		decoded = map[string]any{}
	}
	status, _ := decoded["status"].(string)
	status = strings.ToLower(status)
	if status == "failed" || status == "error" {
		msg, _ := decoded["message"].(string)
		if msg == "" {
			msg = "Request failed"
		}
		return nil, &APIError{Message: msg, HTTPStatus: res.StatusCode, Payload: decoded}
	}
	return decoded, nil
}

func (c *Client) Post(path string, body map[string]any) (map[string]any, error) {
	return c.Request(http.MethodPost, path, body, true)
}

func (c *Client) Get(path string, auth bool) (map[string]any, error) {
	return c.Request(http.MethodGet, path, nil, auth)
}

func (c *Client) CreateWallet(body map[string]any) (map[string]any, error) {
	return c.Post("wallet/create", body)
}

func (c *Client) WalletDetails(body map[string]any) (map[string]any, error) {
	return c.Post("wallet/details", body)
}

func (c *Client) WalletTransactions(body map[string]any) (map[string]any, error) {
	return c.Post("wallet/transactions", body)
}

func (c *Client) InitiateCollection(body map[string]any) (map[string]any, error) {
	return c.Post("wallet/collection/initiate", body)
}

func (c *Client) CollectionInfo(orderReferenceCode string) (map[string]any, error) {
	return c.Post("wallet/collection/info", map[string]any{"order_reference_code": orderReferenceCode})
}

func (c *Client) CreateCollectionMethod(body map[string]any) (map[string]any, error) {
	return c.Post("wallet/collection/method/create", body)
}

func (c *Client) DirectTransfer(body map[string]any) (map[string]any, error) {
	return c.Post("wallet/transfer/direct/single", body)
}

func (c *Client) CreateIndividual(body map[string]any) (map[string]any, error) {
	return c.Post("user/account/individual/create", body)
}

func (c *Client) IndividualDetails(body map[string]any) (map[string]any, error) {
	return c.Post("user/account/individual/details", body)
}

func (c *Client) CreateBusiness(body map[string]any) (map[string]any, error) {
	return c.Post("user/account/business/create", body)
}

func (c *Client) Countries() (map[string]any, error) {
	return c.Get("user/misc/countries", false)
}

func (c *Client) JobTypes() (map[string]any, error) {
	return c.Get("user/misc/job-types", false)
}

func (c *Client) BusinessIndustries() (map[string]any, error) {
	return c.Get("user/misc/business-industry-list", false)
}

func (c *Client) BusinessRegistrationTypes() (map[string]any, error) {
	return c.Get("user/misc/business-registration-type", false)
}

func (c *Client) BusinessTradeTypes() (map[string]any, error) {
	return c.Get("user/misc/business-trade-type", false)
}

func (c *Client) BusinessProductServiceTypes() (map[string]any, error) {
	return c.Get("user/misc/business-product-service-type", false)
}

func (c *Client) SourceOfFundsTypes() (map[string]any, error) {
	return c.Get("user/misc/source-of-funds-type", false)
}

func VerifyWebhook(rawBody, signatureHeader, timestampHeader, apiKey string, toleranceSeconds int) bool {
	if rawBody == "" || signatureHeader == "" || timestampHeader == "" || apiKey == "" {
		return false
	}
	ts, err := strconv.ParseInt(timestampHeader, 10, 64)
	if err != nil {
		return false
	}
	if toleranceSeconds > 0 {
		age := time.Now().Unix() - ts
		if age < 0 {
			age = -age
		}
		if age > int64(toleranceSeconds) {
			return false
		}
	}
	provided := signatureHeader
	if strings.HasPrefix(strings.ToLower(provided), "sha256=") {
		provided = provided[7:]
	}
	mac := hmac.New(sha256.New, []byte(apiKey))
	_, _ = mac.Write([]byte(timestampHeader + "." + rawBody))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(provided))
}

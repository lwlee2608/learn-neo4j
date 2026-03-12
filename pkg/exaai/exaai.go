package exaai

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

const (
	baseURL = "https://api.exa.ai"
	timeout = 30 * time.Second
)

type Config struct {
	ApiKey string `mask:"first=1,last=4"`
}

type ClientImpl struct {
	apiKey     string
	httpClient *http.Client
}

type Client interface {
	Search(ctx context.Context, req SearchRequest) (*SearchResponse, error)
	Answer(ctx context.Context, req AnswerRequest) (*AnswerResponse, error)
}

type SearchRequest struct {
	Query      string   `json:"query"`
	NumResults int      `json:"num_results"`
	Contents   Contents `json:"contents"`
	Category   string   `json:"category,omitempty"`
}

type Contents struct {
	Text    bool     `json:"text"`
	Summary *Summary `json:"summary,omitempty"`
}

type Summary struct {
	Query string `json:"query"`
}

type SearchResponse struct {
	Results []Result `json:"results"`
}

type Result struct {
	ID      string  `json:"id"`
	URL     string  `json:"url"`
	Title   string  `json:"title"`
	Text    string  `json:"text,omitempty"`
	Summary string  `json:"summary,omitempty"`
	Score   float64 `json:"score"`
	Image   string  `json:"image,omitempty"`
}

type AnswerRequest struct {
	Query  string `json:"query"`
	Stream bool   `json:"stream,omitempty"`
	Text   bool   `json:"text,omitempty"`
}

type AnswerResponse struct {
	Answer      string     `json:"answer"`
	Citations   []Citation `json:"citations"`
	CostDollars CostInfo   `json:"costDollars"`
}

type Citation struct {
	ID            string `json:"id"`
	URL           string `json:"url"`
	Title         string `json:"title"`
	Author        string `json:"author,omitempty"`
	PublishedDate string `json:"publishedDate,omitempty"`
	Text          string `json:"text,omitempty"`
	Image         string `json:"image,omitempty"`
	Favicon       string `json:"favicon,omitempty"`
}

type CostInfo struct {
	Total float64 `json:"total"`
}

func NewClient(apiKey string) *ClientImpl {
	return &ClientImpl{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *ClientImpl) Search(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/search", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	return &searchResp, nil
}

func (c *ClientImpl) Answer(ctx context.Context, req AnswerRequest) (*AnswerResponse, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/answer", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var answerResp AnswerResponse
	if err := json.NewDecoder(resp.Body).Decode(&answerResp); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	return &answerResp, nil
}

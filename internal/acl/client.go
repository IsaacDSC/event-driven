package acl

import (
	"bytes"
	"context"
	"encoding/json"
	"event-driven/types"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
	"time"
)

type Client struct {
	httpclient *http.Client
	baseUrl    string
}

// Ensure Client implements ClientInterface
var _ types.ClientInterface = (*Client)(nil)

func NewClient(baseUrl string) *Client {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	return &Client{
		httpclient: client,
		baseUrl:    baseUrl,
	}
}

func (c Client) UpdateInfos(ctx context.Context, txID uuid.UUID, retry int, status string) {
	//TODO: review this responsibility
	var finishedAt time.Time
	if status == "FINISHED" || status == "BACKWARD" || status == "BACKWARD_ERROR" {
		finishedAt = time.Now()
	}

	url := fmt.Sprintf("%s/%s", c.baseUrl, txID)
	input := types.UpdatePayloadInput{
		Status:     status,
		TotalRetry: retry,
		FinishedAt: finishedAt,
	}

	taskInput, err := json.Marshal(input)
	if err != nil {
		log.Printf("could not marshal task: %v\n", err)
	}

	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(taskInput))
	if err != nil {
		fmt.Println(err)
		return
	}
	res, err := c.httpclient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	if res.StatusCode != http.StatusAccepted {
		fmt.Printf("unexpected status code: %d", res.StatusCode)
	}
}

func (c Client) CreateMsg(ctx context.Context, input types.PayloadType) error {
	taskInput, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("could not marshal task: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseUrl, bytes.NewBuffer(taskInput))
	if err != nil {
		return fmt.Errorf("could not create request: %v", err)
	}

	res, err := c.httpclient.Do(req)
	if err != nil {
		return fmt.Errorf("could not send request: %v", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	return nil
}

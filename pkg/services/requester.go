package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	taskParams "fetcher/pkg/api/params/task"
	"fetcher/pkg/api/presenters"
	"fetcher/pkg/services/storage"
)

const httpClientTimeout = 20 * time.Second

type Requester interface {
	Do(params RequesterParams) error
}

type RequesterParams struct {
	TaskKey       string
	RequestParams taskParams.Task
}

type requester struct {
	tempStorageService storage.TempStoreService
}

func NewRequester(tempStorageService storage.TempStoreService) (Requester, error) {
	if tempStorageService == nil {
		return nil, errors.New("services.NewRequester, tempStorageService cannot be empty")
	}

	return &requester{tempStorageService}, nil
}

func (r *requester) Do(params RequesterParams) error {
	// 1. Create request by params
	request, err := http.NewRequest(
		params.RequestParams.Method,
		params.RequestParams.Address,
		bytes.NewBufferString(params.RequestParams.Body))
	if err != nil {
		return err
	}

	for headerKey, headerValue := range params.RequestParams.Headers {
		request.Header.Set(headerKey, headerValue)
	}

	// 2. Create and tune http client
	client := &http.Client{
		Timeout: time.Duration(httpClientTimeout),
	}

	// 3. Finally do request
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	doneResponse := presenters.ClientResponse{
		Key:           params.TaskKey,
		HTTPStatus:    resp.Status,
		Headers:       resp.Header,
		ContentLength: resp.ContentLength,
		Done:          true,
	}

	doneResponseJSON, err := json.Marshal(doneResponse)
	if err != nil {
		return err
	}

	// 4. Store result
	err = r.tempStorageService.Set(params.TaskKey, doneResponseJSON, 0)
	if err != nil {
		return err
	}

	return nil
}

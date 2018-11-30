package taskhandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"fetcher/pkg/api"
	"fetcher/pkg/api/params"
	taskParams "fetcher/pkg/api/params/task"
	"fetcher/pkg/api/presenters"
	"fetcher/pkg/services"
	"fetcher/pkg/services/pool"
	"fetcher/pkg/services/storage"
	"fetcher/pkg/settings"

	"github.com/apsdehal/go-logger"
	"github.com/gorilla/mux"
)

type FetchHandler struct {
	app                services.App
	tempStorageService storage.TempStoreService
	fetchPool          pool.FetchPool
	logger             logger.Logger
}

func New(application services.App,
	tempStorageService storage.TempStoreService,
	fetchPool pool.FetchPool,
	logger logger.Logger) (*FetchHandler, error) {

	if application == nil {
		return nil, errors.New("fetchHandler.New, application cannot be empty")
	}

	if tempStorageService == nil {
		return nil, errors.New("fetchHandler.New, tempStorageService cannot be empty")
	}

	if &logger == nil {
		return nil, errors.New("fetchHandler.New, logger cannot be empty")
	}

	return &FetchHandler{
		app:                application,
		tempStorageService: tempStorageService,
		fetchPool:          fetchPool,
		logger:             logger,
	}, nil
}

func (f *FetchHandler) Fetch(w http.ResponseWriter, req *http.Request) (*api.Response, error) {
	methodParams := taskParams.Task{}

	err := params.MustDecodeParams(req.Body, &methodParams)
	if err != nil {
		f.logger.Criticalf("fetchHandler.Fetch params.MustDecodeParams, unable to parse json params: %s", err)
		return api.InternalServerError(), nil
	}

	validationErrors := params.MustValidateParams(&methodParams)
	if validationErrors != nil {
		return api.ErrorResponse(validationErrors), nil
	}

	// 1. After checks, get last increment value
	lastTaskIncrement, err := f.tempStorageService.GetKeyIncrement(settings.TaskIncrementKey)
	if err != nil {
		f.logger.Criticalf("fetchHandler.Fetch unable to get GetKeyIncrementedId: %s", err)
		return api.InternalServerError(), nil
	}

	if lastTaskIncrement == 0 {
		f.logger.Criticalf("fetchHandler.Fetch GetKeyIncrementedId is 0: %s", err)
		return api.InternalServerError(), nil
	}

	// Task key looks like: "task:1"
	taskKey := storage.TaskKey(lastTaskIncrement)

	// 2. Write key immediately to storage with flag done = false, that the user don't wait.
	notDoneResponse := presenters.ClientResponse{
		Key:  taskKey,
		Done: false,
	}

	notDoneResponseJSON, err := json.Marshal(notDoneResponse)
	if err != nil {
		f.logger.Criticalf("fetchHandler.Fetch unable to marshal notDoneResponse: %s", err)
		return api.InternalServerError(), nil
	}

	err = f.tempStorageService.Set(taskKey, notDoneResponseJSON, 0)
	if err != nil {
		f.logger.Criticalf("fetchHandler.Fetch unable to set notDoneResponseJSON value: %s", err)
		return api.InternalServerError(), nil
	}

	// "Auto increment" mechanism.
	err = f.tempStorageService.Incr(settings.TaskIncrementKey)
	if err != nil {
		f.logger.Criticalf("fetchHandler.Fetch unable to increment taskIncrementKey: %s", err)
		return api.InternalServerError(), nil
	}

	// 3. Add the job to fetch pool worker,
	// when job is finished worker will write response in a store with flag done = true by key.
	workerParams := &services.RequesterParams{
		TaskKey:       taskKey,
		RequestParams: methodParams,
	}

	f.fetchPool.AddWorker(workerParams)

	return api.SuccessResponse(notDoneResponse), nil
}

func (f *FetchHandler) GetTask(w http.ResponseWriter, req *http.Request) (*api.Response, error) {
	vars := mux.Vars(req)
	taskKey := vars["id"]

	if taskKey == "" {
		return api.ErrorResponse(map[string]string{"error": "taskKey not found"}), nil
	}

	task, err := f.tempStorageService.Get(taskKey)
	if err != nil {
		f.logger.Criticalf("fetchHandler.GetTask, unable to get task: %v", err.Error())
		return api.InternalServerError(), nil
	}

	clientResponse := presenters.ClientResponse{}
	err = json.Unmarshal(task, &clientResponse)
	if err != nil {
		f.logger.Criticalf("fetchHandler.GetTask, unable to unmarshal: %v", err.Error())
		return api.InternalServerError(), nil
	}

	return api.SuccessResponse(&clientResponse), nil
}

func (f *FetchHandler) DeleteTask(w http.ResponseWriter, req *http.Request) (*api.Response, error) {
	vars := mux.Vars(req)
	taskKey := vars["id"]

	if taskKey == "" {
		return api.ErrorResponse(map[string]string{"error": "taskKey not found"}), nil
	}

	ok, err := f.tempStorageService.Exist(taskKey)
	if err != nil {
		f.logger.Criticalf("fetchHandler.DeleteTask, unable to check is exist, err: %s", err)
		return api.InternalServerError(), nil
	}

	if !ok {
		return api.NoContentResponse(), nil
	}

	err = f.tempStorageService.Del(taskKey)
	if err != nil {
		f.logger.Criticalf("fetchHandler.DeleteTask, unable to delete task: %s", err)
		return api.InternalServerError(), nil
	}

	return api.EmptySuccessResponse(), nil
}

func (f *FetchHandler) GetTasks(w http.ResponseWriter, req *http.Request) (*api.Response, error) {
	tasks, err := f.tempStorageService.GetAllByPattern(settings.TasksKeyPattern + "*")
	if err != nil {
		f.logger.Criticalf("fetchHandler.GetTasks, unable to get tasks: %s", err)
		return api.InternalServerError(), nil
	}

	var responses []presenters.ClientResponse

	for _, t := range tasks {
		r := &presenters.ClientResponse{}

		err := json.Unmarshal([]byte(t), r)
		if err != nil {
			return nil, fmt.Errorf("redisTempStoreService.GetAllByPattern, json.Unmarshal error: %s", err)
		}

		responses = append(responses, *r)
	}

	return api.SuccessResponse(responses), nil
}

func (f *FetchHandler) GetTasksByPage(w http.ResponseWriter, req *http.Request) (*api.Response, error) {
	vars := mux.Vars(req)

	pageId, err := strconv.Atoi(vars["id"])
	if err != nil {
		f.logger.Errorf("fetchHandler.GetTasksByPage, unable to conver pageId to int: %s", err)
		return api.InternalServerError(), nil
	}

	if pageId <= 0 {
		return api.ErrorResponse(map[string]string{"error": "wrong pageId, pages starts from 1"}), nil
	}

	// Formula for calculating page numbers range
	fromPage := (pageId-1)*settings.TasksPerPageCount + 1
	var requestedTaskKeys []string

	for i := fromPage; i < fromPage+settings.TasksPerPageCount; i++ {
		requestedTaskKeys = append(requestedTaskKeys, settings.TasksKeyPattern+strconv.Itoa(i))
	}

	tasks, err := f.tempStorageService.GetAllByKeys(requestedTaskKeys...)
	if err != nil {
		f.logger.Criticalf("fetchHandler.GetTasks, unable to get tasks: %s", err)
		return api.InternalServerError(), nil
	}

	var responses []presenters.ClientResponse

	for _, t := range tasks {
		r := &presenters.ClientResponse{}

		err := json.Unmarshal([]byte(t), r)
		if err != nil {
			return nil, fmt.Errorf("redisTempStoreService.GetAllByPattern, json.Unmarshal error: %s", err)
		}

		responses = append(responses, *r)
	}

	return api.SuccessResponse(responses), nil
}

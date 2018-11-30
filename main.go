package main

import (
	"fmt"
	"net/http"

	"fetcher/pkg/api"
	"fetcher/pkg/api/taskhandler"
	"fetcher/pkg/services"
	"fetcher/pkg/services/pool"
	"fetcher/pkg/services/storage"

	"github.com/gorilla/mux"
)

func main() {
	app := services.Application()

	err := app.Init()
	if err != nil {
		panic(fmt.Sprintf("app init error: %s", err))
	}

	redisPool := app.RedisPool()

	tempStorageService, err := storage.NewRedisStoreService(redisPool)
	if err != nil {
		panic(fmt.Sprintf("redis store init error: %s", err))
	}

	requester, err := services.NewRequester(tempStorageService)
	if err != nil {
		panic(fmt.Sprintf("requester init error: %s", err))
	}

	fetchPool := pool.NewFetchPool(requester)
	go fetchPool.Start()

	taskHandler, err := taskhandler.New(app, tempStorageService, fetchPool, app.Logger())
	if err != nil {
		panic(fmt.Sprintf("task handler init error: %s", err))
	}

	r := mux.NewRouter()

	r.HandleFunc("/task/fetch", api.ResponseHandler(taskHandler.Fetch)).Methods("POST")
	r.HandleFunc("/task/{id}", api.ResponseHandler(taskHandler.GetTask)).Methods("GET")
	r.HandleFunc("/task/{id}", api.ResponseHandler(taskHandler.DeleteTask)).Methods("DELETE")
	r.HandleFunc("/tasks", api.ResponseHandler(taskHandler.GetTasks)).Methods("GET")
	r.HandleFunc("/tasks/page/{id}", api.ResponseHandler(taskHandler.GetTasksByPage)).Methods("GET")

	fmt.Println("listen")
	http.ListenAndServe(":8080", r)
}

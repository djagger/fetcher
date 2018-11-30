# Fetcher

This service allows to execute HTTP requests to other resources in the worker pool and store results.

### API Endpoints:

`POST /task/fetch` allow the service to perform an HTTP request to a specific resource.

For example:
```
curl -X POST \
  http://localhost:8080/task/fetch \
  -H 'Content-Type: application/json' \
  -d '{
	"method": "GET",
	"address": "http://ozon.ru/context/detail/id/200",
	"headers": {
		"Content-Type": "text/html;charset=windows-1251",
		"Allow": "GET",
		"Content-Length": "1984"
	},
	"body": "some text"
}'
```

All requests are performed by a special pool worker in the goroutines.
Pool worker settings (number of workers, buffer size) located in _pkg/services/pool/fetch_pool.go_.

`GET /task/{id}` returns task by _id_.

`DELETE /task/{id}` removes task by _id_.

`GET /tasks` returns all tasks from storage.

`GET /tasks/page/{id}` returns tasks by pages

The value of the count of tasks per one page is stored in the _TasksPerPageCount_ in _pkg/settings/task.go_.

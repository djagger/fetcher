package storage

import (
	"strconv"

	"fetcher/pkg/settings"
)

func TaskKey(id int64) string {
	return settings.TasksKeyPattern + strconv.FormatInt(id, 10)
}

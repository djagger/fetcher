package storage

import "strconv"

const taskKeyPrefix = "task:"

func TaskKey(id int64) string {
	return taskKeyPrefix + strconv.FormatInt(id, 10)
}

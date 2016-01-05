package main

import (
	"testing"
)

func TestTaskSyncFacebookFriends(t *testing.T) {

	task := &SyncFacebookFriends{
		UserId:  15918606474806281,
		Fbid:    "116857838692042",
		Fbtoken: "CAAMn9Gs6TKoBAD5Q49snWZCIiIBXtr972pWHzdEtmrxvBQimS0NhPEJnAauAM7vK6vRLoljd0Vufv7IuR9UBzcPmBQJmfF4F3iuzzUv2AewnC856cG1mvULn4wVifdhCGAGHAmxs3rNALNrD1ZBvicttNd9eSL57ZBUnglm0OYvJ4ZBhUpycSyLBmbYObbtWqOmMnJQpJQZDZD",
	}

	task.Run(server.task_executor)
}

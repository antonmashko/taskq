package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/antonmashko/taskq"
	"github.com/go-redis/redis/v9"
)

type Task struct {
	ID int
}

func (t *Task) Do(ctx context.Context) error {
	fmt.Println("task ID:", t.ID)
	return nil
}

// Using Redis List https://redis.io/docs/data-types/lists/ for implement queue
type RedisQueue struct {
	key string
	rdb *redis.Client
}

func (q *RedisQueue) Enqueue(ctx context.Context, task taskq.Task) (int64, error) {
	t, ok := task.(*Task)
	if !ok {
		return -1, errors.New("invalid task type")
	}
	b, err := json.Marshal(t)
	if err != nil {
		return -1, err
	}
	return q.rdb.RPush(ctx, q.key, b).Result()
}

func (q *RedisQueue) Dequeue(ctx context.Context) (taskq.Task, error) {
	// NOTE:
	// If process stops during executing task, task can/will be lost
	// use some intermediate lists for storing tasks that are in progress

	str, err := q.rdb.LPop(ctx, q.key).Result()
	if err != nil {
		if err == redis.Nil {
			// IMPORTANT: return `taskq.EmptyQueue` error if queue is empty
			// this is signal for TaskQ for stopping Dequeue process
			return nil, taskq.EmptyQueue
		}
		return nil, err
	}

	// Pay attention, the task will be unmarshalled from json. All private fields and pointers on complex structs will be lost
	var result Task
	err = json.Unmarshal([]byte(str), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

/*
	Run example:
	1. start redis container:
	$ docker run -d --rm -p 6379:6379 --name redis redis
	2. run example
	$ go run main.go

	Output:
	task ID: 10
	task ID: 11
	task ID: 12
	task ID: 13
	task ID: 14
	task ID: 15
	task ID: 16
	task ID: 17
	task ID: 18
	task ID: 19

*/

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: ":6379",
	})

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Fatalln("ping failed.", err)
	}

	tq := taskq.NewWithQueue(0, &RedisQueue{
		key: "taskq_redis_example",
		rdb: rdb,
	})
	// Start will also restore tasks execution if some of tasks will be in redis queue
	tq.Start()

	for i := 10; i < 20; i++ {
		// enqueue tasks to redis queue
		_, err := tq.Enqueue(context.Background(), &Task{ID: i})
		if err != nil {
			log.Fatalf("enqueue failed. i=%d err=%s", i, err)
		}
	}

	// waiting until all tasks will be finished
	err := tq.Shutdown(taskq.ContextWithWait(context.Background()))
	if err != nil {
		log.Fatalln("shutdown failed.", err)
	}
}

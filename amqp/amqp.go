package amqp

import (
	"encoding/json"
	b "github.com/pmurley/mida/base"
	"github.com/streadway/amqp"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type ConnParams struct {
	User string
	Pass string
	Host string
	Port int
	Tls  bool
}

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	tag     string
	done    chan error
}

// LoadTasks handles loading MIDA tasks in to AMQP (probably RabbitMQ) queue.
func LoadTasks(tasks b.TaskSet, params ConnParams, queue string, priority uint8, shuffle bool) (int, error) {
	rabbitUri := "amqp://" + params.User + ":" + params.Pass + "@" + params.Host + ":" + strconv.Itoa(params.Port)
	if params.Tls {
		rabbitUri = strings.Replace(rabbitUri, "amqp://", "amqps://", 1)
	}

	connection, err := amqp.Dial(rabbitUri)
	if err != nil {
		return 0, err
	}
	defer connection.Close()

	channel, err := connection.Channel()
	if err != nil {
		return 0, err
	}

	if shuffle {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(tasks),
			func(i, j int) { tasks[i], tasks[j] = tasks[j], tasks[i] })
	}

	tasksLoaded := 0
	for _, task := range tasks {
		taskBytes, err := json.Marshal(task)
		if err != nil {
			return tasksLoaded, err
		}

		err = channel.Publish(
			"",
			queue,
			false,
			false,
			amqp.Publishing{
				Headers:      amqp.Table{},
				ContentType:  "text/plain",
				DeliveryMode: 0,
				Priority:     priority,
				Timestamp:    time.Now(),
				Body:         taskBytes,
			})
		if err != nil {
			return tasksLoaded, err
		}

		// Successfully loaded a task into the queue
		tasksLoaded += 1
	}

	return tasksLoaded, nil
}

package queue

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	commonTypes "github.com/revan730/clipper-common/types"
	"github.com/streadway/amqp"
)

// CIQueue is a name of rabbitMQ queue with CI jobs
const CIQueue = "ciJobs"

// CDQueue is a name of rabbitMQ queue with CD jobs
const CDQueue = "cdJobs"

// Queue is used for operations with rabbitMQ queues
type Queue struct {
	rabbitConnection *amqp.Connection
	channel          *amqp.Channel
	CIJobsQueue      amqp.Queue
	CDJobsQueue      amqp.Queue
}

func declareQueue(ch *amqp.Channel, queueName string) (amqp.Queue, error) {
	return ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
}

// NewQueue creates new copy of Queue
func NewQueue(addr string) *Queue {
	conn, err := amqp.Dial(addr)
	if err != nil {
		panic(fmt.Sprintf("Couldn't connect to rabbitmq: %s", err))
	}
	ch, err := conn.Channel()
	if err != nil {
		panic(fmt.Sprintf("Couldn't open rabbitmq channel: %s", err))
	}
	ciQueue, err := declareQueue(ch, CIQueue)
	if err != nil {
		panic(fmt.Sprintf("Couldn't declare CI jobs queue: %s", err))
	}
	cdQueue, err := declareQueue(ch, CDQueue)
	if err != nil {
		panic(fmt.Sprintf("Couldn't declare CD jobs queue: %s", err))
	}
	Queue := &Queue{
		rabbitConnection: conn,
		channel:          ch,
		CIJobsQueue:      ciQueue,
		CDJobsQueue:      cdQueue,
	}

	return Queue
}

// publish sends protobuf encoded message to provided queue
func (q *Queue) publish(msg proto.Message, queue string) error {
	body, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return q.channel.Publish(
		"", queue, false, false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
}

// PublishCDJob publishes CDJob with provided data
func (q *Queue) PublishCDJob(jobMsg *commonTypes.CDJob) error {
	return q.publish(jobMsg, q.CDJobsQueue.Name)
}

// makeMsgChan creates channel to retranslate message bodies to
func (q *Queue) makeMsgChan(queue string) (chan []byte, error) {
	incomingMsgs, err := q.channel.Consume(queue, "", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	msgsChan := make(chan []byte)
	go func() {
		for m := range incomingMsgs {
			msgsChan <- m.Body
		}
	}()
	return msgsChan, nil
}

// MakeCIMsgChan creates channel with raw CI job messages
func (q *Queue) MakeCIMsgChan() (chan []byte, error) {
	return q.makeMsgChan(q.CIJobsQueue.Name)
}

// Close gracefully breaks connection with rabbitMQ
func (q *Queue) Close() {
	q.rabbitConnection.Close()
}

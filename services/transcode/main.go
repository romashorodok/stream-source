package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
	"time"

	transcodetopicpb "github.com/romashorodok/stream-source/pb/go/kafka/topic/v1"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"google.golang.org/protobuf/proto"
)

const (
	KAFKA = "localhost:9092"
)

var config = &kafka.ConfigMap{
	"bootstrap.servers": KAFKA,
	"group.id":          "upload-consumers",
	"auto.offset.reset": "earliest",
}

var signals = make(chan os.Signal, 1)

type ConsumerContainer[F proto.Message] struct {
	Message kafka.Message
	Data    F
}

func NewConsumerContainer[F proto.Message]() *ConsumerContainer[F] {
	container := &ConsumerContainer[F]{}
	typeOfdata := reflect.TypeOf(container.Data).Elem()
	data := reflect.New(typeOfdata).Interface().(proto.Message)
	container.Data = data.(F)
	return container
}

func consumeProtobufTopic[F proto.Message](topic string, out chan<- *ConsumerContainer[F]) {
	for {
		log.Println("Consumer iteration")

		c, err := kafka.NewConsumer(config)
		if err != nil {
			log.Printf("Error creating consumer: %v\n", err)
			time.Sleep(20 * time.Second)
			continue
		}

		if err = c.Subscribe(topic, nil); err != nil {
			log.Printf("Error subscribing to topic: %v\n", err)
			time.Sleep(20 * time.Second)
			continue
		}

		defer c.Close()

		for {
			log.Println("Consumer reader iteration")

			container := NewConsumerContainer[F]()

			msg, err := c.ReadMessage(-1)

			if err != nil {
				switch e := err.(type) {
				case kafka.Error:
					switch e.Code() {
					case kafka.ErrTimedOut:
						log.Println("Timed out while waiting for message")

					case kafka.ErrTransport:
						log.Println("Connection to broker lost")

					default:
						// %4|1681132527.138|MAXPOLL|rdkafka#consumer-1| [thrd:main]: Application maximum poll interval (300000ms) exceeded by 162ms (adjust max.poll.interval.ms for long-running message processing): leaving group
						// 2023/04/10 16:15:57 Error reading message: Application maximum poll interval (300000ms) exceeded by 162ms
						// NOTE: if consume it loong time, if somehow handle it, if do that may escape one for loop
						log.Printf("Error reading message: %v\n", err)
					}
				default:
					log.Printf("Error reading message: %v\n", err)
				}

				c.Close()

				break
			}

			if err = proto.Unmarshal(msg.Value, container.Data); err != nil {
				// TODO: When process failed, topic must be published back
				log.Printf("Error unmarshalling message: %v", err)
				continue
			}

			container.Message = *msg

			log.Printf("Delivered message to topic %s [%d] at offset %v\n",
				*msg.TopicPartition.Topic, msg.TopicPartition.Partition, msg.TopicPartition.Offset)

			// FIX: when chan has 1 size it read 1 topic but if chan is full it wait until can write topic into chan

			out <- container
		}
	}
}

func processTopicMessages(ctx context.Context, topicChan <-chan *ConsumerContainer[*transcodetopicpb.TranscodeAudio], numWorkers int) {
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	workerPool := make(chan struct{}, numWorkers)
	for i := 0; i < numWorkers; i++ {
		workerPool <- struct{}{}
	}

	for {
		log.Println("Worker iteration")

		select {
		case <-ctx.Done():
			wg.Wait()
			return
		case msg := <-topicChan:
			<-workerPool
			go func(msg *ConsumerContainer[*transcodetopicpb.TranscodeAudio]) {
				defer func() {
					workerPool <- struct{}{}
					wg.Done()
					wg.Add(1)
				}()

				// log.Println("Start processing")
				// log.Println(msg.Data)
				//
				// log.Printf("Processing message to topic %s [%d] at offset %v\n",
				// 	*msg.Message.TopicPartition.Topic, msg.Message.TopicPartition.Partition, msg.Message.TopicPartition.Offset)
				//
				time.Sleep(time.Second * 30)
				// log.Println("End processing message")
			}(msg)
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	containerChan := make(chan *ConsumerContainer[*transcodetopicpb.TranscodeAudio])
	go consumeProtobufTopic(transcodetopicpb.Default_TranscodeAudio_Topic, containerChan)
	go processTopicMessages(ctx, containerChan, 1)

	select {
	case <-signals:
		log.Println("Termination signal received")
	case <-ctx.Done():
		log.Println("Context cancelled")
	}
}

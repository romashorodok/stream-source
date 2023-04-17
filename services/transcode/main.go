package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	transcodetopicpb "github.com/romashorodok/stream-source/pb/go/kafka/topic/v1"
	"github.com/romashorodok/stream-source/pkgs/consumer"
	"github.com/romashorodok/stream-source/services/transcode/transcoder"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
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

func processTopicMessages(ctx context.Context, topicChan <-chan *consumer.ConsumerBox[*transcodetopicpb.TranscodeAudio], numWorkers int) {
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": KAFKA})

	if err != nil {
		log.Panic(err)
	}

	defer p.Close()

	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					log.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	transcodesvc := transcoder.NewTranscoderService(&ctx, p)

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
			go func(msg *consumer.ConsumerBox[*transcodetopicpb.TranscodeAudio]) {
				defer func() {
					workerPool <- struct{}{}
					wg.Done()
					wg.Add(1)
				}()
				log.Printf("Processing message to topic %s [%d] at offset %v\n",
					*msg.Message.TopicPartition.Topic, msg.Message.TopicPartition.Partition, msg.Message.TopicPartition.Offset)

				transcodesvc.TranscodeAudio(&transcoder.TranscodeData{
					Bucket:     *msg.Data.Bucket,
					BucketId:   *msg.Data.BucketId,
					OriginFile: *msg.Data.OriginFile,
				})

				log.Printf("End pocessing message to topic %s [%d] at offset %v\n",
					*msg.Message.TopicPartition.Topic, msg.Message.TopicPartition.Partition, msg.Message.TopicPartition.Offset)
			}(msg)
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	containerChan := make(chan *consumer.ConsumerBox[*transcodetopicpb.TranscodeAudio])
	go consumer.ConsumeProtobufTopic(config, transcodetopicpb.Default_TranscodeAudio_Topic, containerChan)
	go processTopicMessages(ctx, containerChan, 4)

	select {
	case <-signals:
		log.Println("Termination signal received")
	case <-ctx.Done():
		log.Println("Context cancelled")
	}
}

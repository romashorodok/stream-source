package workers

import (
	"context"
	"log"
	"sync"

	"github.com/google/uuid"
	audiotopicpb "github.com/romashorodok/stream-source/pb/go/topic/audio/v1"
	"github.com/romashorodok/stream-source/pkgs/consumer"
	"github.com/romashorodok/stream-source/services/audio/types"
	"gorm.io/gorm"
)

type TranscodedAudioWorker struct {
	DB *gorm.DB
}

type transcodedAudioResult struct {
	BucketId     uuid.UUID
	Bucket       string
	ManifestFile string
}

func (s *TranscodedAudioWorker) updateAudioBucketManifest(transcodedAudio *transcodedAudioResult) {
	var bucket types.AudioBucket

	s.DB.Where("audio_bucket_id = ?", transcodedAudio.BucketId).
		First(&bucket)

	bucket.Manifest = transcodedAudio.ManifestFile

	s.DB.Updates(bucket)
}

func (s *TranscodedAudioWorker) ProcessTranscodedAudio(ctx context.Context, data <-chan *consumer.ConsumerBox[*audiotopicpb.AudioTranscoded], numWorkers int) {
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	workerPool := make(chan struct{}, numWorkers)
	for i := 0; i < numWorkers; i++ {
		workerPool <- struct{}{}
	}

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return
		case msg := <-data:
			<-workerPool

			go func(msg *consumer.ConsumerBox[*audiotopicpb.AudioTranscoded]) {
				defer func() {
					workerPool <- struct{}{}
					wg.Done()
					wg.Add(1)
				}()

				log.Printf("Processing message to topic %s [%d] at offset %v\n",
					*msg.Message.TopicPartition.Topic, msg.Message.TopicPartition.Partition, msg.Message.TopicPartition.Offset)

				transcoded := &transcodedAudioResult{
					Bucket:       *msg.Data.Bucket,
					ManifestFile: *msg.Data.ManifestFile,
				}

				if id, err := uuid.Parse(*msg.Data.BucketId); err == nil {
					transcoded.BucketId = id
				}

				s.updateAudioBucketManifest(transcoded)

				log.Printf("End pocessing message to topic %s [%d] at offset %v\n",
					*msg.Message.TopicPartition.Topic, msg.Message.TopicPartition.Partition, msg.Message.TopicPartition.Offset)
			}(msg)
		}
	}
}

package types

import (
	"github.com/google/uuid"
	audiopb "github.com/romashorodok/stream-source/pb/go/audio/v1"
)

type Audio struct {
	AudioId uuid.UUID `gorm:"primarykey;type:uuid;default:gen_random_uuid()"`

	AudioBucketId *uuid.UUID   `gorm:"type:uuid"`
	AudioBucket   *AudioBucket `gorm:"foreignKey:AudioId;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`

	Title string
}

func (m *Audio) Proto() *audiopb.Audio {
	return &audiopb.Audio{
		AudioId: m.AudioId.String(),
		Title:   m.Title,
	}
}

func (m *Audio) FromProto(model *audiopb.Audio) *Audio {
	if id, err := uuid.FromBytes([]byte(model.GetAudioId())); err == nil {
		m.AudioId = id
	}
	m.Title = model.GetTitle()
	return m
}

type AudioBucket struct {
	AudioBucketId uuid.UUID `gorm:"primarykey;type:uuid;default:gen_random_uuid()"`

	AudioId *uuid.UUID `gorm:"type:uuid;null"`
	// Audio   *Audio     `gorm:"foreignKey:AudioBucketId;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`

	Bucket     string
	OriginFile string `gorm:"null"`
	Manifest   string `gorm:"null"`
}

func (m *AudioBucket) Proto() *audiopb.AudioBucket {
	return &audiopb.AudioBucket{
		AudioBucketId: m.AudioBucketId.String(),
		Bucket:        m.Bucket,
		OriginFile:    m.OriginFile,
	}
}
func (m *AudioBucket) FromProto(model *audiopb.AudioBucket) *AudioBucket {
	if id, err := uuid.Parse(model.GetAudioBucketId()); err == nil {
		m.AudioBucketId = id
	}

	m.OriginFile = model.GetOriginFile()
	m.Bucket = model.GetBucket()
	return m
}

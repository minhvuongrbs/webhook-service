package kafka

import (
	"fmt"

	"github.com/IBM/sarama"
	"github.com/spf13/cast"
)

const (
	TraceIDHeaderKey = "trace_id"
	CreatedAtKey     = "created_at"
	AttemptTimesKey  = "attempt_times"
	OrderingKey      = "ordering_key"

	FinUUIDHeaderKey   = "fin_uuid"
	FinCreatedAtKey    = "fin_created_at"
	FinExecutedAtKey   = "fin_executed_at"
	FinAttemptTimesKey = "fin_attempt_times"
)

type Metadata map[string]string

func (m Metadata) Get(key string) string {
	if v, ok := m[key]; ok {
		return v
	}

	return ""
}

func (m Metadata) Set(key, value string) {
	m[key] = value
}

type ConsumerMessage struct {
	//for ordering kafka message, the same key is will be in the same partition
	//Example using order_id for ordering message of a order
	OrderingKey string `json:"ordering_key"`

	// TraceID is an unique identifier of message. it will attach to header of kafka message
	TraceID string `json:"trace_id"`

	// Metadata contains the message metadata. it will attach to header of kafka message
	Metadata Metadata `json:"metadata"`

	// Payload is message's payload. the data after marshalling
	Payload []byte `json:"payload"`

	// CreatedAt is the time when message, it will attach to header of kafka message, unix milisecond
	CreatedAt int64 `json:"created_at"`

	// AttemptTimes use for retry logic, with CreatedTime. number of retries, start from 0
	AttemptTimes int `json:"attempt_times"`

	//Topic is the topic of kafka message
	Topic string `json:"topic"`

	// ConsumerMessage information, use for consumer, get from kafka message
	Partition int32 `json:"partition"`
	Offset    int64 `json:"offset"`
}

func (m *ConsumerMessage) GetOrderingKey() string {
	if m == nil {
		return ""
	}
	return m.OrderingKey
}

func (m *ConsumerMessage) GetTopic() string {
	if m == nil {
		return ""
	}
	return m.Topic
}

func (m *ConsumerMessage) GetTraceID() string {
	if m == nil {
		return ""
	}
	return m.TraceID
}

func (m *ConsumerMessage) GetPartition() int32 {
	if m == nil {
		return 0
	}
	return m.Partition
}

func (m *ConsumerMessage) GetOffset() int64 {
	if m == nil {
		return 0
	}
	return m.Offset
}

func (m *ConsumerMessage) GetCreatedAt() int64 {
	if m == nil {
		return 0
	}
	return m.CreatedAt
}

func (m *ConsumerMessage) GetAttemptTimes() int {
	if m == nil {
		return 0
	}
	return m.AttemptTimes
}

func (m *ConsumerMessage) GetMetadata() Metadata {
	if m == nil {
		return nil
	}
	metadata := make(Metadata, len(m.Metadata))
	for key, value := range m.Metadata {
		metadata[key] = value
	}
	return metadata
}

func (m *ConsumerMessage) GetPayload() []byte {
	if m == nil {
		return nil
	}
	return m.Payload
}

func (m *ConsumerMessage) Clone() ConsumerMessage {
	result := ConsumerMessage{}
	if m == nil {
		return result
	}
	result.OrderingKey = m.OrderingKey
	result.TraceID = m.TraceID
	result.Metadata = m.GetMetadata()
	result.Payload = m.Payload
	result.CreatedAt = m.CreatedAt
	result.AttemptTimes = m.AttemptTimes
	result.Topic = m.Topic
	result.Partition = m.Partition
	result.Offset = m.Offset
	return result
}

func Unmarshal(kafkaMsg *sarama.ConsumerMessage) *ConsumerMessage {
	var traceID, orderingKey string
	var createdAt int64
	var attemptTimes int
	metadata := make(Metadata, len(kafkaMsg.Headers))

	for _, header := range kafkaMsg.Headers {
		switch string(header.Key) {
		case TraceIDHeaderKey, FinUUIDHeaderKey:
			traceID = string(header.Value)
		case CreatedAtKey, FinCreatedAtKey:
			createdAt = cast.ToInt64(string(header.Value))
		case AttemptTimesKey, FinAttemptTimesKey:
			attemptTimes = cast.ToInt(string(header.Value))
		case OrderingKey:
			orderingKey = string(header.Value)
		default:
			metadata.Set(string(header.Key), string(header.Value))
		}
	}

	msg := &ConsumerMessage{
		TraceID:      traceID,
		OrderingKey:  orderingKey,
		CreatedAt:    createdAt,
		AttemptTimes: attemptTimes,
		Topic:        kafkaMsg.Topic,
		Partition:    kafkaMsg.Partition,
		Offset:       kafkaMsg.Offset,
		Metadata:     metadata,
		Payload:      kafkaMsg.Value,
	}
	return msg
}

type ProducerMessage struct {
	//for ordering kafka message, the same key is will be in the same partition
	//Example using order_id for ordering message of a order
	OrderingKey string

	// TraceID is an unique identifier of message. it will attach to header of kafka message
	TraceID string

	// Metadata contains the message metadata. it will attach to header of kafka message
	Metadata Metadata

	// Payload is message's payload. the data after marshalling
	Payload []byte

	// CreatedAt is the time when message, it will attach to header of kafka message, unix milisecond
	CreatedAt int64

	// AttemptTimes use for retry logic, with CreatedTime. number of retries
	AttemptTimes int
}

func (m *ProducerMessage) GetAttemptTimes() int {
	if m == nil {
		return 0
	}
	return m.AttemptTimes
}

func (m *ProducerMessage) GetCreatedAt() int64 {
	if m == nil {
		return 0
	}
	return m.CreatedAt
}

func (m *ProducerMessage) GetTraceID() string {
	if m == nil {
		return ""
	}
	return m.TraceID
}

func (m *ProducerMessage) GetOrderingKey() string {
	if m == nil {
		return ""
	}
	return m.OrderingKey
}

func Marshal(m *ProducerMessage) *sarama.ProducerMessage {
	headers := []sarama.RecordHeader{
		{
			Key:   []byte(TraceIDHeaderKey),
			Value: []byte(m.TraceID),
		},
		{
			Key:   []byte(CreatedAtKey),
			Value: []byte(fmt.Sprintf("%d", m.CreatedAt)),
		},
		{
			Key:   []byte(AttemptTimesKey),
			Value: []byte(fmt.Sprintf("%d", m.AttemptTimes)),
		},
		{
			Key:   []byte(OrderingKey),
			Value: []byte(m.OrderingKey),
		},
		{
			Key:   []byte(FinUUIDHeaderKey),
			Value: []byte(m.TraceID),
		},
		{
			Key:   []byte(FinCreatedAtKey),
			Value: []byte(fmt.Sprintf("%d", m.CreatedAt)),
		},
		{
			Key:   []byte(FinAttemptTimesKey),
			Value: []byte(fmt.Sprintf("%d", m.AttemptTimes)),
		},
	}

	for key, value := range m.Metadata {
		headers = append(headers, sarama.RecordHeader{
			Key:   []byte(key),
			Value: []byte(value),
		})
	}

	return &sarama.ProducerMessage{
		Key:     sarama.StringEncoder(m.OrderingKey), //Ordering kafka message
		Value:   sarama.ByteEncoder(m.Payload),
		Headers: headers,
	}
}

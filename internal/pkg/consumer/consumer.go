package consumer

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/maximus335/utm_postback_middle/internal/app/utm_postback_middle/config"
	"github.com/maximus335/utm_postback_middle/internal/pkg/events"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type Consumer struct {
	kafcaConfig *config.KafkaConfiguration
	apiConfig   *config.AppsflyerConfiguration
	logger      *logrus.Logger
	db          *pgxpool.Pool
}

type message struct {
	ApplicationUid string `json:"application_guid"`
	Status         string `json:"status"`
}

func New(kafcaConfig *config.KafkaConfiguration, apiConfig *config.AppsflyerConfiguration, logger *logrus.Logger, db *pgxpool.Pool) *Consumer {
	return &Consumer{
		kafcaConfig: kafcaConfig,
		apiConfig:   apiConfig,
		logger:      logger,
		db:          db,
	}
}

func (c *Consumer) StartKafkaReaders() {
	for _, topic := range c.kafcaConfig.Topics {
		reader := c.getKafkaReader(topic)
		c.logger.Info("Consumer Started ", topic)
		go c.runReadMessages(reader)
	}
}

func (c *Consumer) getKafkaReader(topic string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.kafcaConfig.Brokers,
		Topic:    topic,
		GroupID:  "postback-middle",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
}

func (c *Consumer) runReadMessages(reader *kafka.Reader) {
	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			c.logger.Error(err, " ", reader.Config().Topic)
			continue
		}
		c.logger.Info("is a message ", string(m.Value))
		msg, err := c.fetchMessage(m.Value)
		if err != nil {
			c.logger.Error(err)
			continue
		}
		if msg.Status == "active" {
			err = events.CreateEventFromKafka(c.db, msg.Status, msg.ApplicationUid, c.apiConfig)
			if err != nil {
				c.logger.Error("Event from kafka didn`t created, error: ", err)
			}
		}
	}
}

func (c *Consumer) fetchMessage(data []byte) (*message, error) {
	var msg message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

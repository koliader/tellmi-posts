package rabbitmq

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
)

func (c *Client) ConsumeUpdateUser(handler func(req UserUpdated) error) error {
	messages, err := c.GetMessages(UserUpdatedQueue)
	if err != nil {
		return fmt.Errorf("error to get update user queue messages: %v", err)
	}
	go func() {
		for message := range messages {
			var messageBody UserUpdated
			err := json.Unmarshal(message.Body, &messageBody)
			if err != nil {
				log.Error().Err(err).Msg("error to unmarshal rabbitmq message")
				continue
			}

			err = handler(messageBody)
			if err != nil {
				log.Error().Err(err).Msg("error to handle rabbitmq message")
				continue
			}
			log.Info().Msg("user updated in posts successfully")
		}
	}()
	return nil
}

func (c *Client) ConsumeUserCreated(handler func(req UserCreated) error) error {
	messages, err := c.GetMessages(UserCreatedQueue)
	if err != nil {
		return fmt.Errorf("error to get user created queue messages: %v", err)
	}
	go func() {
		for message := range messages {
			var messageBody UserCreated
			err := json.Unmarshal(message.Body, &messageBody)
			if err != nil {
				log.Error().Err(err).Msg("error to unmarshal rabbitmq message")
				continue
			}

			err = handler(messageBody)
			if err != nil {
				log.Error().Err(err).Msg("error to handle rabbitmq message")
				continue
			}
			log.Info().Msg("user created event processed in posts successfully")
		}
	}()
	return nil
}

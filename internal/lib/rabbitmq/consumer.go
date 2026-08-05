package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	sdkrabbitmq "github.com/koliader/tellmi-sdk/rabbitmq"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// classify decides whether a handler failure is permanent (reject, no
// requeue) or transient (requeue). Duplicate/missing-user outcomes from the
// idempotent services are permanent: retrying them can never succeed.
func classify(err error) error {
	switch status.Code(err) {
	case codes.AlreadyExists, codes.NotFound, codes.InvalidArgument, codes.PermissionDenied:
		return fmt.Errorf("%w: %v", sdkrabbitmq.ErrReject, err)
	default:
		return err
	}
}

func ConsumeUpdateUser(ctx context.Context, client *sdkrabbitmq.Client, handler func(req sdkrabbitmq.UserUpdated) error) error {
	return client.Consume(ctx, sdkrabbitmq.UserUpdatedQueue, func(body []byte) error {
		var messageBody sdkrabbitmq.UserUpdated
		if err := json.Unmarshal(body, &messageBody); err != nil {
			return fmt.Errorf("%w: %v", sdkrabbitmq.ErrReject, err)
		}
		if err := handler(messageBody); err != nil {
			return classify(err)
		}
		log.Info().Msg("user updated in posts successfully")
		return nil
	})
}

func ConsumeUserCreated(ctx context.Context, client *sdkrabbitmq.Client, handler func(req sdkrabbitmq.UserCreated) error) error {
	return client.Consume(ctx, sdkrabbitmq.UserCreatedQueue, func(body []byte) error {
		var messageBody sdkrabbitmq.UserCreated
		if err := json.Unmarshal(body, &messageBody); err != nil {
			return fmt.Errorf("%w: %v", sdkrabbitmq.ErrReject, err)
		}
		if err := handler(messageBody); err != nil {
			return classify(err)
		}
		log.Info().Msg("user created event processed in posts successfully")
		return nil
	})
}

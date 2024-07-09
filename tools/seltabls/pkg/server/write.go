package server

import (
	"context"
	"fmt"
	"io"

	"github.com/conneroisu/seltabl/tools/seltabls/pkg/rpc"
)

// ResponseWriter is an interface for writing a response
type ResponseWriter interface {
	WriteResponse(
		ctx context.Context,
		writer *io.Writer,
		msg rpc.MethodActor,
	) error
}

// WriteResponse writes a message to the writer
func WriteResponse(
	ctx context.Context,
	writer *io.Writer,
	msg rpc.MethodActor,
) error {
	for {
		if ctx.Err() != nil {
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		default:
			reply, err := rpc.EncodeMessage(msg)
			if err != nil {
				return fmt.Errorf(
					"failed to encode response to request (%s): %w",
					msg.Method(),
					err,
				)
			}
			res, err := (*writer).Write([]byte(reply))
			if err != nil {
				return fmt.Errorf(
					"failed to encode response to request (%s): %w",
					msg.Method(),
					err,
				)
			}
			if res != len(reply) {
				return fmt.Errorf(
					"failed writing all of response to (%s) request: %w",
					msg.Method(),
					err,
				)
			}
			return nil
		}
	}
}

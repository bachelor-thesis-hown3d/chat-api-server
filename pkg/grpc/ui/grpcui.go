package ui

import (
	"context"
	"log"
	"net/http"

	"github.com/fullstorydev/grpcui/standalone"
	"google.golang.org/grpc"
)

func NewServer(ctx context.Context, target string, logger *log.Logger) error {
	cc, err := grpc.DialContext(ctx, target, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		return err
	}
	h, err := standalone.HandlerViaReflection(ctx, cc, target)
	if err != nil {
		return err
	}
	logger.Printf("Serving grpc webui on port: %v", 9090)
	return http.ListenAndServe(":9090", h)
}

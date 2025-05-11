package server

import (
	"context"
	"log"

	"github.com/437d5/subpub/pkg/pb"
	"github.com/437d5/subpub/pkg/subpub"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	pb.UnimplementedPubSubServer
	Subpub       subpub.SubPub
	ShutdownChan chan struct{}
}

func (s Server) Subscribe(req *pb.SubscribeRequest, stream grpc.ServerStreamingServer[pb.Event]) error {
	ctx := stream.Context()

	var sub subpub.Subscription
	defer func() {
		if sub != nil {
			sub.Unsubscribe()
		}
	}()

	var err error
	log.Println("New subscription")
	sub, err = s.Subpub.Subscribe(req.Key, func(msg any) {
		select {
		case <-ctx.Done():
			return
		case <-s.ShutdownChan:
			return
		default:
			if data, ok := msg.(string); ok {
				if err := stream.Send(&pb.Event{
					Data: data,
				}); err != nil {
					log.Printf("failed to send message: %v", err)
				}
				log.Println("New message sended")
			}
		}
	})

	if err != nil {
		return status.Errorf(codes.Internal, "subsctibe failed: %v", err)
	}

	select {
	case <-ctx.Done():
	case <-s.ShutdownChan:
	}

	return nil
}

func (s Server) Publish(ctx context.Context, req *pb.PublishRequest) (*emptypb.Empty, error) {
	log.Println("New message published")
	if err := s.Subpub.Publish(req.Key, req.Data); err != nil {
		return nil, status.Errorf(codes.Internal, "publish failed: %v", err)
	}

	return &emptypb.Empty{}, nil
}

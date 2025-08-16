package grpcserver

import (
	"context"
	"fmt"
	"log"
	"net/http"

	connect "connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	logsv1 "github.com/plasma-containers/plasma/gen/logs/v1"
	"github.com/plasma-containers/plasma/gen/logs/v1/logsv1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// TODO: configure address and/or port
const address = "localhost:8081"

type loggerServiceServer struct {
	logsv1connect.UnimplementedLoggerServiceHandler
}

func (s *loggerServiceServer) LogStream(
	ctx context.Context,
	req *connect.Request[logsv1.LogStreamRequest],
	stream *connect.ServerStream[logsv1.LogStreamResponse],
) error {
	name := req.Msg.GetName()
	log.Printf("Got a request for logs from container %s", name)
	for i := 1; i <= 5; i++ {
		if err := stream.Send(&logsv1.LogStreamResponse{
			Message: fmt.Sprintf("Message %d", i)}); err != nil {
			return err
		}
	}
	return nil
}

func Run() {
	mux := http.NewServeMux()
	path, handler := logsv1connect.NewLoggerServiceHandler(&loggerServiceServer{})
	mux.Handle(path, handler)
	// TODO: probably best to disable reflection on non-dev deployment
	reflector := grpcreflect.NewStaticReflector(
		logsv1connect.LoggerServiceName,
	)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	log.Println("oOoOo gRPC listening on", address, "oOoOo")
	http.ListenAndServe(
		address,
		// Use h2c so we can serve HTTP/2 without TLS.
		h2c.NewHandler(mux, &http2.Server{}),
	)
}

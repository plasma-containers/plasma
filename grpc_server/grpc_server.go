package grpcserver

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"

	connect "connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"github.com/pgulb/plasma/container"
	logsv1 "github.com/pgulb/plasma/gen/logs/v1"
	"github.com/pgulb/plasma/gen/logs/v1/logsv1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const address = "0.0.0.0:8081"

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
	c := make(chan container.LogResult, 1)
	buffer := bytes.Buffer{}
	scanner := bufio.NewScanner(&buffer)
	go container.GoLogs(name, c)
	for result := range c {
		// if channel returns error
		if result.Err != nil {
			var errWithMark []byte
			errWithMark = fmt.Appendf(errWithMark, "........<grpc-error> %s\n", result.Err.Error())
			if errSend := stream.Send(&logsv1.LogStreamResponse{
				Message: errWithMark}); errSend != nil {
				return errSend
			}
		}
		// if channel returns log data
		buffer.Write(result.Value)
		// if log line did not end yet
		if !bytes.Contains(result.Value, []byte{'\n'}) {
			continue
		}
		if scanner.Scan() {
			if err := stream.Send(&logsv1.LogStreamResponse{
				Message: scanner.Bytes()}); err != nil {
				return err
			}
		}
	}
	// read rest of log lines if any
	for scanner.Scan() {
		log.Println("Sending some remaining log line from", name)
		if err := stream.Send(&logsv1.LogStreamResponse{
			Message: scanner.Bytes()}); err != nil {
			return err
		}
	}
	log.Println("Done sending logs from", name)
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

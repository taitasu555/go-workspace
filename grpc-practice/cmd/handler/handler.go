package handler

import (
	"context"
	"errors"
	"fmt"
	hellopb "grpc-practice/pkg/grpc/api"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type myServer struct {
	hellopb.UnimplementedGreetingServiceServer
}

func NewHandler() hellopb.GreetingServiceServer {
	return &myServer{}
}

// Unary RPC 1リクエスト-1レスポンス
func (s *myServer) Hello(ctx context.Context, req *hellopb.HelloRequest) (*hellopb.HelloResponse, error) {
	// リクエストからnameフィールドを取り出して
	// "Hello, [名前]!"というレスポンスを返す

	// stat := status.New(codes.Unknown, "unknown error")
	// stat, _ = stat.WithDetails(&errdetails.DebugInfo{
	// 	Detail: "debug info for error",
	// })

	// err := stat.Err()

	headerMD := metadata.New(map[string]string{"type": "unary", "from": "server", "in": "header"})

	if err := grpc.SetHeader(ctx, headerMD); err != nil {
		return nil, err
	}

	trailerMD := metadata.New(map[string]string{"type": "unary", "from": "server", "in": "trailer"})
	if err := grpc.SetTrailer(ctx, trailerMD); err != nil {
		return nil, err
	}

	// if md, ok := metadata.FromIncomingContext(ctx); ok {
	// 	fmt.Println("metadata: ", md)
	// }

	return &hellopb.HelloResponse{
		Message: fmt.Sprintf("Hello, %s!", req.GetName()),
	}, nil

	// return &hellopb.HelloResponse{}, err
}

// ストリーム処理
func (s *myServer) HelloServerStream(req *hellopb.HelloRequest, stream hellopb.GreetingService_HelloServerStreamServer) error {
	resCount := 5

	for i := 0; i < resCount; i++ {
		// streamのSendメソッドを使っている
		if err := stream.Send(&hellopb.HelloResponse{
			Message: fmt.Sprintf("[%d] Hello, %s!", i, req.GetName()),
		}); err != nil {
			return err
		}
		time.Sleep(time.Second * 1)
	}

	// return文でメソッドを終了させる=ストリームの終わり
	return nil
}

func (s *myServer) HelloClientStream(stream hellopb.GreetingService_HelloClientStreamServer) error {
	nameList := make([]string, 0)

	for {
		req, err := stream.Recv()

		if errors.Is(err, io.EOF) {
			message := fmt.Sprintf("Hello, %v!", nameList)
			return stream.SendAndClose(&hellopb.HelloResponse{
				Message: message,
			})
		}

		if err != nil {
			return err
		}

		nameList = append(nameList, req.GetName())
	}
}

func (s *myServer) HelloBiStreams(stream hellopb.GreetingService_HelloBiStreamsServer) error {
	if md, ok := metadata.FromIncomingContext(stream.Context()); ok {
		log.Println(md)
	}
	for {
		// 1. リクエスト受信
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		message := fmt.Sprintf("Hello, %v!", req.GetName())
		if err := stream.Send(&hellopb.HelloResponse{
			Message: message,
		}); err != nil {
			return err
		}
	}
}

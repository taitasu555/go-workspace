package server

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"grpc-practice/cmd/handler"
	hellopb "grpc-practice/pkg/grpc/api"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func Server() {
	port := 8080

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		panic(err)
	}

	//grpc serverを作成
	s := grpc.NewServer()

	helloHandler := handler.NewHandler()

	//grpcにサーバーを登録
	hellopb.RegisterGreetingServiceServer(s, helloHandler)

	reflection.Register(s)

	go func() {
		log.Printf("gRPC server is running!")
		s.Serve(listener)
	}()

	quit := make(chan os.Signal, 1)
	/*
	 *signal.Notify 関数を使って、os.Interrupt シグナル（通常、Ctrl+C を押したときに発生するシグナル）を quit チャネルに送る
	 */
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("stopping gRPC server...")
	s.GracefulStop()
}

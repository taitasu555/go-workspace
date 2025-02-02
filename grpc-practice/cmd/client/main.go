package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"os"

	"grpc-practice/cmd/client/interceptor"
	hellopb "grpc-practice/pkg/grpc/api"

	//Detailsメソッドにて取得したメッセージ型をデシリアライズして中身を見るために、
	//errdetailsパッケージをimportする必要がある
	_ "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	scanner *bufio.Scanner
	client  hellopb.GreetingServiceClient
)

func main() {
	fmt.Println("start grpc client")

	scanner = bufio.NewScanner(os.Stdin)

	address := "localhost:8080"

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithUnaryInterceptor(interceptor.MyUnaryClientInteceptor1), grpc.WithStreamInterceptor(interceptor.MyStreamClientInteceptor1))

	if err != nil {
		log.Fatal("connection error: ", err)
		return
	}

	defer conn.Close()

	client = hellopb.NewGreetingServiceClient(conn)

	for {
		fmt.Println("1: send Request")
		fmt.Println("2: HelloServerStream")
		fmt.Println("3: HelloClientStream")
		fmt.Println("4: HelloBiStream")
		fmt.Println("5: exit")
		fmt.Print("please enter >")

		scanner.Scan()
		in := scanner.Text()

		switch in {
		case "1":
			Hello()

		case "2":
			HelloServerStream()

		case "3":
			HelloClientStream()

		case "4":
			HelloBiStreams()

		case "5":
			fmt.Println("bye.")
			goto M
		}
	}

M:
}

func Hello() {
	fmt.Println("Please enter your name.")
	var header, trailer metadata.MD
	scanner.Scan()
	name := scanner.Text()

	req := &hellopb.HelloRequest{
		Name: name,
	}

	ctx := context.Background()
	md := metadata.New(map[string]string{"type": "unary", "from": "client"})
	ctx = metadata.NewOutgoingContext(ctx, md)

	res, err := client.Hello(ctx, req, grpc.Header(&header), grpc.Trailer(&trailer))
	if err != nil {
		fmt.Println(err)
		if status, ok := status.FromError(err); ok {
			fmt.Printf("code: %s\n", status.Code())
			fmt.Printf("message: %s\n", status.Message())
			fmt.Printf("details: %s\n", status.Details())
		} else {
			fmt.Println(err)
		}
	} else {
		fmt.Println(header)
		fmt.Println(trailer.Get("type"))
		fmt.Println(res.GetMessage())
	}
}

func HelloServerStream() {
	fmt.Println("please enter your name")

	scanner.Scan()
	name := scanner.Text()

	req := &hellopb.HelloRequest{
		Name: name,
	}

	//HelloServerStreamメソッドを呼んで、サーバーからレスポンスが送られてくる
	stream, err := client.HelloServerStream(context.Background(), req)

	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		//そのストリームのRecvメソッドを呼ぶことでレスポンスを得る
		res, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("all the responses have already received.")
			break
		}

		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(res)
	}
}

func HelloClientStream() {
	//ライアントが持つHelloClientStreamメソッドを呼んで、サーバーからリクエストを送るストリーム(GreetingService_HelloClientStreamClientインターフェース型)を取得
	stream, err := client.HelloClientStream(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	sendCount := 5
	fmt.Printf("Please enter %d names.\n", sendCount)
	for i := 0; i < sendCount; i++ {
		scanner.Scan()
		name := scanner.Text()

		//Sendメソッドを、HelloRequest型の引数と共に呼び出すことでリクエストを送信
		if err := stream.Send(&hellopb.HelloRequest{
			Name: name,
		}); err != nil {
			fmt.Println(err)
			return
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(res.GetMessage())
	}
}

func HelloBiStreams() {
	ctx := context.Background()
	md := metadata.New(map[string]string{"type": "stream", "from": "client"})
	ctx = metadata.NewOutgoingContext(ctx, md)
	stream, err := client.HelloBiStreams(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	sendNum := 5
	fmt.Printf("Please enter %d names.\n", sendNum)

	var sendEnd, recvEnd bool
	sendCount := 0
	for !(sendEnd && recvEnd) {
		// 送信処理
		if !sendEnd {
			scanner.Scan()
			name := scanner.Text()

			sendCount++
			if err := stream.Send(&hellopb.HelloRequest{
				Name: name,
			}); err != nil {
				fmt.Println(err)
				sendEnd = true
			}

			if sendCount == sendNum {
				sendEnd = true
				if err := stream.CloseSend(); err != nil {
					fmt.Println(err)
				}
			}
		}

		// 受信処理
		if !recvEnd {
			if res, err := stream.Recv(); err != nil {
				if !errors.Is(err, io.EOF) {
					fmt.Println(err)
				}
				recvEnd = true
			} else {
				fmt.Println(res.GetMessage())
			}
		}
	}
}

package main

import (
	"context"
	"fmt"
	"github.com/naraycitra/grpc-go-adventure/internal/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"net"
	"time"
)

type server struct {
}

func (*server) Greet(ctx context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	fmt.Printf("Greet service was invoked with %v\n", req)
	fn := req.GetGreeting().GetFirstName()
	ln := req.GetGreeting().GetLastName()
	result := "Hello, " + fn + ln
	res := &greetpb.GreetResponse{
		Result: result,
	}
	return res, nil
}

func (*server) GreetManyTimes(req *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer) error {
	fmt.Printf("GreetManyTimes service was invoked with %v\n", req)
	fn := req.GetGreeting().GetFirstName()
	ln := req.GetGreeting().GetLastName()

	for i := 0; i < 10; i++ {
		result := fmt.Sprintf("Hello, %v %v at %d", fn, ln, i)
		res := &greetpb.GreetManyTimesResponse{
			Result: result,
		}
		stream.Send(res)
		time.Sleep(1 * time.Second)
	}
	return nil
}

func (*server) LongGreet(stream greetpb.GreetService_LongGreetServer) error {
	val := &stream
	fmt.Printf("LongGreet service was invoked with %v\n", *val)
	var result string
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&greetpb.LongGreetResponse{
				Result: result,
			})
		}
		if err != nil {
			log.Fatalf("Error received stream from client: %v", err)
			return err
		}
		fn := req.GetGreeting().GetFirstName()
		ln := req.GetGreeting().GetLastName()
		result += "Hello, " + fn + " " + ln + "! \n"
	}
}

func (*server) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error {
	val := &stream
	fmt.Printf("LongGreet service was invoked with %v\n", *val)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error received stream from client: %v", err)
			return err
		}
		fn := req.GetGreeting().GetFirstName()
		ln := req.GetGreeting().GetLastName()
		result := fmt.Sprintf("Hello, %s %s", fn, ln)
		res := &greetpb.GreetEveryoneResponse{
			Result: result,
		}
		err = stream.Send(res)
		if err != nil {
			log.Fatalf("Error while send stream to client: %v", err)
			return err
		}
	}
}

func (*server) GreetWithDeadLine(ctx context.Context, req *greetpb.GreetWithDeadLineRequest) (*greetpb.GreetWithDeadLineResponse, error) {
	fmt.Printf("GreetWithDeadLine service was invoked with %v\n", req)
	for i := 0; i < 3; i++ {
		if ctx.Err() == context.Canceled {
			// the client canceled the request
			fmt.Println("The Client cancel the request")
			return nil, status.Error(codes.Canceled, "The Client canceled the request")
		}
		time.Sleep(1 * time.Second)
	}
	fn := req.GetGreeting().GetFirstName()
	ln := req.GetGreeting().GetLastName()
	result := "Hello, " + fn + ln
	res := &greetpb.GreetWithDeadLineResponse{
		Result: result,
	}
	return res, nil
}

func main() {
	fmt.Println("Greet Server started...")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")

	if err != nil {
		log.Fatalf("Failed to listen:%v", err)
	}

	s := grpc.NewServer()
	greetpb.RegisterGreetServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve:%v", err)
	}

}

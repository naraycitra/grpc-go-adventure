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
	"time"
)

func main() {
	fmt.Println("Greet Client")

	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial server:%v", err)
	}
	defer cc.Close()

	c := greetpb.NewGreetServiceClient(cc)

	//fmt.Printf("Created client %v", c)
	doUnary(c)

	doServerStreaming(c)

	doClientStreaming(c)

	doClientServerStreaming(c)

	doUnaryWithDeadLine(c, 1*time.Second)

	doUnaryWithDeadLine(c, 5*time.Second)

}

func doUnary(c greetpb.GreetServiceClient) {
	fmt.Println("Starting a unary RPC")
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Naray",
			LastName:  "Citra",
		},
	}
	res, err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling Geet RPC: %v", err)
	}
	log.Printf("Response from Greet :%v", res.Result)
}

func doServerStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting a streaming server RPC")
	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Naray",
			LastName:  "Citra",
		},
	}
	stream, err := c.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling RPC:%v", err)
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			// stream reach end of file
			break
		}
		if err != nil {
			log.Fatalf("Error while streaming :%v", err)
		}
		log.Printf("Response from server: %v", res.GetResult())
	}
}

func doClientStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting a streaming client RPC")
	requests := []*greetpb.LongGreetRequest{
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Naray",
				LastName:  "Citra",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Nurul",
				LastName:  "Istiqomah",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Abdullah",
				LastName:  "Muhammad AlFatih",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Fatimah",
				LastName:  "Azzahra Khawla",
			},
		},
	}
	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("error while calling LongGreet: %v", err)
	}
	for _, req := range requests {
		fmt.Printf("Sending stream request: %v \n", req)
		stream.Send(req)
		time.Sleep(100 * time.Millisecond)
	}
	result, err := stream.CloseAndRecv()
	fmt.Printf("Result from server:\n %v", result.GetResult())
}

func doClientServerStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting a streaming client-server RPC")
	requests := []*greetpb.GreetEveryoneRequest{
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Naray",
				LastName:  "Citra",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Nurul",
				LastName:  "Istiqomah",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Abdullah",
				LastName:  "Muhammad AlFatih",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Fatimah",
				LastName:  "Azzahra Khawla",
			},
		},
	}
	// we create a stream by invoking the client

	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("error while calling LongGreet: %v", err)
	}
	wc := make(chan struct{})
	// we send bunch of messages to the server (go routine)
	go func() {
		// function for sending message
		for _, req := range requests {
			fmt.Printf("Sending stream request: %v \n", req)
			stream.Send(req)
			time.Sleep(100 * time.Millisecond)
		}
		stream.CloseSend()
	}()
	// we receive a bunch of message from the server (go routine)
	go func() {
		// function for receiving message
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				// stream reach end of file
				break
			}
			if err != nil {
				log.Fatalf("Error while streaming :%v", err)
			}
			log.Printf("Response from server: %v", res.GetResult())
		}
		close(wc)
	}()
	// block until everything is done
	<-wc
}

func doUnaryWithDeadLine(c greetpb.GreetServiceClient, timeout time.Duration) {
	fmt.Println("Starting a unary with deadline RPC")
	req := &greetpb.GreetWithDeadLineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Naray",
			LastName:  "Citra",
		},
	}
	ctx, _ := context.WithTimeout(context.Background(), timeout)

	res, err := c.GreetWithDeadLine(ctx, req)
	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				fmt.Printf("Timeout was hit! Deadline has exceeded")
			} else {
				fmt.Printf("unexpected error: %v", statusErr)
			}
		} else {
			log.Fatalf("Error while calling Geet RPC: %v", err)
		}
		return
	}
	log.Printf("Response from GreetWithDeadLine :%v", res.Result)
}

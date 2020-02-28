package main

import (
	"context"
	"fmt"
	"github.com/naraycitra/grpc-go-adventure/internal/calculator/calculatorpb"
	"google.golang.org/grpc"
	"io"
	"log"
	"time"
)

func main() {
	fmt.Println("Calculator Client")

	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial server:%v", err)
	}
	defer cc.Close()

	c := calculatorpb.NewCalculatorServiceClient(cc)

	//doCalculate(c)
	//
	//doPrimeNumberDecomposition(c)

	//doComputeAverage(c)

	//doFindMaximum(c)

	doSquareRoot(c)
}

func doCalculate(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting calculate x and y")
	req := &calculatorpb.CalculatorRequest{
		Calculating: &calculatorpb.Calculating{
			X: 3,
			Y: 10,
		},
	}
	res, err := c.Calculate(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling Calculate RPC: %v", err)
	}
	log.Printf("Sumary of %v,%v is:%v", req.Calculating.GetX(), req.Calculating.GetY(), res.Result)
}

func doPrimeNumberDecomposition(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting PrimeNumberDecomposition")
	req := &calculatorpb.PrimeNumberDecompositionRequest{
		Number: 120,
	}
	stream, err := c.PrimeNumberDecomposition(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling RPC: %v", err)
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
		log.Printf("Response from server: %v\n", res.GetResult())
	}
}

func doComputeAverage(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting ComputeAverage")
	requests := []*calculatorpb.ComputeAverageRequest{
		{
			Number: 1,
		},
		{
			Number: 2,
		},
		{
			Number: 3,
		},
		{
			Number: 4,
		},
	}
	stream, err := c.ComputeAverage(context.Background())
	if err != nil {
		log.Fatalf("error while calling ComputeAverage: %v", err)
	}
	for _, req := range requests {
		fmt.Printf("Sending stream request: %v \n", req)
		stream.Send(req)
		time.Sleep(100 * time.Millisecond)
	}
	result, err := stream.CloseAndRecv()
	fmt.Printf("Result from server:\n %v", result.GetResult())
}

func doFindMaximum(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting FindMaximum")

	requests := []int64{1, 5, 3, 6, 2, 20}

	// we create a stream by invoking the client

	stream, err := c.FindMaximum(context.Background())
	if err != nil {
		log.Fatalf("error while calling FindMaximum: %v", err)
	}
	wc := make(chan struct{})
	// we send bunch of messages to the server (go routine)
	go func() {
		// function for sending message
		for _, req := range requests {
			fmt.Printf("Sending stream request: %v \n", req)
			stream.Send(&calculatorpb.FindMaximumRequest{
				Number: req,
			})
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

func doSquareRoot(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting calculate x and y")
	req := &calculatorpb.SquareRootRequest{
		Number: -1,
	}
	res, err := c.SquareRoot(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling Calculate RPC: %v", err)
	}
	log.Printf("SquareRoot of %v is:%v", req.GetNumber(), res.GetRoot())
}

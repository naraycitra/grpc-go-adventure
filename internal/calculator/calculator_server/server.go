package main

import (
	"context"
	"fmt"
	"github.com/naraycitra/grpc-go-adventure/internal/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"math"
	"net"
)

type server struct {
}

func (*server) Calculate(ctx context.Context, req *calculatorpb.CalculatorRequest) (*calculatorpb.CalculatorResponse, error) {
	fmt.Printf("Calculate Service invoked with: %v", req)
	x := req.GetCalculating().GetX()
	y := req.GetCalculating().GetY()
	sum := x + y
	result := &calculatorpb.CalculatorResponse{
		Result: sum,
	}
	return result, nil
}

func (*server) PrimeNumberDecomposition(req *calculatorpb.PrimeNumberDecompositionRequest, stream calculatorpb.CalculatorService_PrimeNumberDecompositionServer) error {
	fmt.Printf("PrimeNumberDecomposition Service invoked with: %v", req)
	var k int64
	k = 2
	n := req.GetNumber()
	for n > 1 {
		if n%k == 0 {
			res := &calculatorpb.PrimeNumberDecompositionResponse{
				Result: k,
			}
			stream.Send(res)
			n /= k
		} else {
			k++
			fmt.Printf("Divisor has incrased by:%v", k)
		}
	}
	return nil
}

func (*server) ComputeAverage(stream calculatorpb.CalculatorService_ComputeAverageServer) error {
	val := &stream
	fmt.Printf("ComputeAverage service was invoked with %v\n", *val)
	var result float64
	i := 0
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			result /= float64(i)
			return stream.SendAndClose(&calculatorpb.ComputeAverageResponse{
				Result: result,
			})
		}
		if err != nil {
			log.Fatalf("Error received stream from client: %v", err)
		}
		result += float64(req.GetNumber())
		i++
	}
}

func (*server) FindMaximum(stream calculatorpb.CalculatorService_FindMaximumServer) error {
	val := &stream
	fmt.Printf("ComputeAverage service was invoked with %v\n", *val)
	var result int64
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error received stream from client: %v", err)
			return err
		}
		n := req.GetNumber()
		if n > result {
			result = n
		}
		res := &calculatorpb.FindMaximumResponse{
			Result: result,
		}
		err = stream.Send(res)
		if err != nil {
			log.Fatalf("Error while send stream to client: %v", err)
			return err
		}
	}
}

func (*server) SquareRoot(ctx context.Context, req *calculatorpb.SquareRootRequest) (*calculatorpb.SquareRootResponse, error) {
	number := req.GetNumber()
	if number < 0 {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Received negative number: %v", number))
	}
	return &calculatorpb.SquareRootResponse{
		Root: math.Sqrt(float64(number)),
	}, nil
}

func main() {
	fmt.Println("Calcutator Server starting...")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()

	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve:%v", err)
	}

}

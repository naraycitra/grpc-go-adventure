#!/bin/bash

protoc internal/greet/greetpb/greet.proto --go_out=plugins=grpc:.
protoc internal/calculator/calculatorpb/calculator.proto --go_out=plugins=grpc:.
protoc internal/blog/blogpb/blog.proto --go_out=plugins=grpc:.
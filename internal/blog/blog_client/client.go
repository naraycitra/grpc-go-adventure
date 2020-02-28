package main

import (
	"context"
	"github.com/naraycitra/grpc-go-adventure/internal/blog/blogpb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"io"
)

var sugar *zap.SugaredLogger

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	sugar = logger.Sugar()

	sugar.Info("Blog client")
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		sugar.Errorf("Error while dialing server: %v\n", err)
	}
	defer cc.Close()

	c := blogpb.NewBlogServiceClient(cc)

	doCreateBlog(c)

	//doReadBlog(c)

	//doUpdateBlog(c)
	//doDeleteBlog(c)
	doReadListBlog(c)
}

func doCreateBlog(client blogpb.BlogServiceClient) {
	sugar.Info("Create blog...")
	req := &blogpb.CreateBlogRequest{
		Blog: &blogpb.Blog{
			AuthorId: "naraycitra",
			Title:    "Golang is cool",
			Content:  "Go lang is really cool programming language",
		},
	}
	res, err := client.CreateBlog(context.Background(), req)

	if err != nil {
		sugar.Errorf("Failed to create blog: %v", err)
	}
	sugar.Infof("Created blog: %v", res.GetBlog())
}

func doReadBlog(client blogpb.BlogServiceClient) {
	sugar.Info("Read Blog")

	// scenario that blog not exists
	_, err := client.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: "123456df"})

	if err != nil {
		sugar.Errorf("Error when reading: %v", err)
	}

	// scenario that blog exists
	req := &blogpb.ReadBlogRequest{
		BlogId: "5e5498017b8f574de2770ba6",
	}
	res, err := client.ReadBlog(context.Background(), req)
	if err != nil {
		sugar.Error("Error when reading: %v", err)
	}
	data := &blogpb.ReadBlogResponse{
		Blog: &blogpb.Blog{
			Id:       res.GetBlog().GetId(),
			AuthorId: res.GetBlog().GetAuthorId(),
			Title:    res.GetBlog().GetTitle(),
			Content:  res.GetBlog().GetContent(),
		},
	}
	sugar.Error("Blog read: %v\n", data)
}

func doUpdateBlog(client blogpb.BlogServiceClient) {
	sugar.Info("Update Blog")

	// update Blog
	// scenario that blog not exists
	_, err := client.UpdateBlog(context.Background(), &blogpb.UpdateBlogRequest{
		Blog: &blogpb.Blog{
			Id:       "123456df",
			AuthorId: "Changed Author",
			Title:    "My First Blog (edited)",
			Content:  "Content of the first blog, with some awesome additions!",
		},
	})

	if err != nil {
		sugar.Errorf("Error when update: %v", err)
	}

	// scenario that blog exists
	updated := &blogpb.UpdateBlogRequest{
		Blog: &blogpb.Blog{
			Id:       "5e5498017b8f574de2770ba6",
			AuthorId: "Changed Author",
			Title:    "My First Blog (edited)",
			Content:  "Content of the first blog, with some awesome additions!",
		},
	}
	res, err := client.UpdateBlog(context.Background(), updated)
	if err != nil {
		sugar.Errorf("Error when update: %v", err)
	}
	data := &blogpb.UpdateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       res.GetBlog().GetId(),
			AuthorId: res.GetBlog().GetAuthorId(),
			Title:    res.GetBlog().GetTitle(),
			Content:  res.GetBlog().GetContent(),
		},
	}
	sugar.Infof("Blog update: %v", data)
}

func doDeleteBlog(client blogpb.BlogServiceClient) {
	sugar.Info("Delete Blog")

	// delete Blog
	// scenario that blog not exists
	_, err := client.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{
		BlogId: "123456df",
	})

	if err != nil {
		sugar.Errorf("Error when delete: %v", err)
	}

	// scenario that blog exists
	deleted := &blogpb.DeleteBlogRequest{
		BlogId: "5e5498017b8f574de2770ba6",
	}
	res, err := client.DeleteBlog(context.Background(), deleted)
	if err != nil {
		sugar.Errorf("Error when update: %v", err)
	}

	sugar.Infof("Blog update: %v", res)
}

func doReadListBlog(client blogpb.BlogServiceClient) {
	sugar.Info("Read list of Blog")

	stream, err := client.ListBlog(context.Background(), &blogpb.ListBlogRequest{})

	if err != nil {
		sugar.Errorf("error while streaming ListBlog RPC:%v", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			sugar.Info("End of blog list")
			break
		}

		if err != nil {
			sugar.Errorf("something not right happened:%v", err)
		}
		sugar.Infof("Blog: %v", res.GetBlog())
	}

}

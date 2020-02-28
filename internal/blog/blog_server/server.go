package main

import (
	"context"
	"fmt"
	"github.com/naraycitra/grpc-go-adventure/internal/blog/blogpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
	"os"
	"os/signal"
	"time"
)

var collection *mongo.Collection
var sugar *zap.SugaredLogger

type server struct {
}

type blogItem struct {
	objectID primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Title    string             `bson:"title"`
	Content  string             `bson:"content"`
}

func (*server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	sugar.Infof("Create Blog: %s", req.GetBlog().GetTitle())
	blog := req.GetBlog()

	data := blogItem{
		AuthorID: blog.GetAuthorId(),
		Title:    blog.GetTitle(),
		Content:  blog.GetContent(),
	}

	res, err := collection.InsertOne(context.Background(), data)

	if err != nil {
		sugar.Errorf("Error while insert data: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("cannot insert data:%v/n", err))
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)

	if !ok {
		sugar.Error("cannot convert oid")
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("cannot convert oid/n"))
	}
	return &blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       oid.Hex(),
			AuthorId: blog.GetAuthorId(),
			Title:    blog.GetTitle(),
			Content:  blog.GetContent(),
		},
	}, nil
}

func (*server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	sugar.Infof("Read blog: %s", req.GetBlogId())
	blogID := req.GetBlogId()

	oid, err := primitive.ObjectIDFromHex(blogID)

	if err != nil {
		sugar.Errorf("Error while parse ObjectID from Hex:%v", err)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("cannot parse ID: %v/n", err))
	}

	// create empty struct
	blogData := &blogItem{}

	filter := bson.M{"_id": oid}

	res := collection.FindOne(context.Background(), filter)

	if err := res.Decode(blogData); err != nil {
		sugar.Errorf("Cannot find blog with ID:%v", err)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Cannot find blog with ID:%v/n", err))
	}

	return &blogpb.ReadBlogResponse{
		Blog: dataToBlogPb(blogData),
	}, nil
}

func dataToBlogPb(data *blogItem) *blogpb.Blog {
	return &blogpb.Blog{
		Id:       data.objectID.Hex(),
		AuthorId: data.AuthorID,
		Title:    data.Title,
		Content:  data.Content,
	}
}

func (*server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	sugar.Infof("Update blog: %v/n", req.GetBlog())
	blog := req.GetBlog()

	oid, err := primitive.ObjectIDFromHex(blog.GetId())
	if err != nil {
		sugar.Errorf("cannot parse id: %s", blog.GetId())
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("cannot parse id/n"))
	}

	// create emtpy struct
	data := &blogItem{}
	filter := bson.M{"_id": oid}

	res := collection.FindOne(context.Background(), filter)

	if err := res.Decode(data); err != nil {
		sugar.Errorf("cannot find blog with specified id:%v", err)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("cannot find blog with specified id:%v/n", err))
	}

	// we update internal struct
	data.AuthorID = blog.GetAuthorId()
	data.Title = blog.GetTitle()
	data.Content = blog.GetContent()

	ur, err := collection.ReplaceOne(context.Background(), filter, data)
	sugar.Infof("Update result:%v", ur)

	if err != nil {
		sugar.Errorf("cannot update object in MongoDB: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("cannot update object in MongoDB:%v/n", err))
	}

	return &blogpb.UpdateBlogResponse{
		Blog: dataToBlogPb(data),
	}, nil
}

func (*server) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	sugar.Infof("Delete blog: %v/n", req.GetBlogId())
	blog := req.GetBlogId()

	oid, err := primitive.ObjectIDFromHex(blog)
	if err != nil {
		sugar.Errorf("cannot parse id: %s", blog)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("cannot parse id/n"))
	}

	filter := bson.M{"_id": oid}

	ur, err := collection.DeleteOne(context.Background(), filter)

	if err != nil {
		sugar.Errorf("cannot delete object in MongoDB: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("cannot delete object in MongoDB:%v/n", err))
	}

	if ur.DeletedCount == 0 {
		sugar.Errorf("cannot delete object in MongoDB: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("cannot delete object in MongoDB:%v/n", err))
	}

	return &blogpb.DeleteBlogResponse{
		BlogId: req.GetBlogId(),
	}, nil
}

func (*server) ListBlog(req *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer) error {
	sugar.Info("List blog request")

	cur, err := collection.Find(context.Background(), primitive.D{{}})
	if err != nil {
		sugar.Errorf("unknown internal error: %v", err)
		return status.Errorf(codes.Internal, fmt.Sprintf("unknown internal error: %v", err))
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		data := &blogItem{}
		err := cur.Decode(data)
		if err != nil {
			sugar.Errorf("error while decoding data: %v", err)
			return status.Errorf(codes.Internal, fmt.Sprintf("error while decoding data: %v", err))
		}
		stream.Send(&blogpb.ListBlogResponse{
			Blog: dataToBlogPb(data),
		})
	}
	if err := cur.Err(); err != nil {
		sugar.Errorf("unknown error: %v", err)
		return status.Errorf(codes.Internal, fmt.Sprintf("unknown error: %v", err))
	}
	return nil
}

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	sugar = logger.Sugar()

	sugar.Info("Blog Service starting")

	sugar.Info("Connecting to mongodb...")
	// connect to mongodb
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	collection = client.Database("mydb").Collection("blog")

	if err != nil {
		sugar.Errorf("failed connect to database: %v", err)
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		sugar.Errorf("Failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	s := grpc.NewServer(opts...)
	blogpb.RegisterBlogServiceServer(s, &server{})
	go func() {
		sugar.Info("starting the server...")
		if err := s.Serve(lis); err != nil {
			sugar.Errorf("error to server: %v", err)
		}
	}()

	// Wait ctrl+C to stop the server
	ch := make(chan os.Signal, 1)

	signal.Notify(ch, os.Interrupt)

	// Block until signal ctrl+C received
	<-ch
	sugar.Info("Stopping the server")
	s.Stop()
	sugar.Info("Closing the Listener")
	lis.Close()
	sugar.Info("Disconnecting from mongodb")
	client.Disconnect(context.TODO())
	sugar.Info("Blog service stopped")
}

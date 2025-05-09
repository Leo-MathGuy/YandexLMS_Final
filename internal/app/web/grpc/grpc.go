package grpc

import (
	"context"
	"net"

	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/logging"
	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/storage"
	pb "github.com/Leo-MathGuy/YandexLMS_Final/internal/grpc"
	"google.golang.org/grpc"
)

type taskServer struct {
	pb.UnimplementedTasksServer
}

func (s *taskServer) GetTask(ctx context.Context, req *pb.Empty) (*pb.TaskData, error) {
	task := storage.GetReadyTask(&storage.T)

	if task == nil {
		return &pb.TaskData{Have: false}, nil
	} else {
		return &pb.TaskData{Id: uint32(task.ID), Left: *task.Left, Right: *task.Right, Operator: string(*task.Op), Have: true}, nil
	}
}

func (s *taskServer) SubmitTask(ctx context.Context, req *pb.TaskSubmit) (*pb.Empty, error) {
	return &pb.Empty{}, storage.FinishTask(&storage.T, uint(req.Id), req.Result)
}

type GRPCControl struct {
	GrpcServer *grpc.Server
	Listener   net.Listener
}

func StartServer(address string) *GRPCControl {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		logging.Panic("Error starting GRPC: %s", err.Error())
		return nil
	}

	grpcServer := grpc.NewServer()
	pb.RegisterTasksServer(grpcServer, &taskServer{})

	go func() {
		logging.Log("gRPC server listening on %s", address)
		if err := grpcServer.Serve(lis); err != nil {
			logging.Error("gRPC server error: %v", err)
		}
	}()

	return &GRPCControl{
		GrpcServer: grpcServer,
		Listener:   lis,
	}
}

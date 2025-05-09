package grpc

import (
	"context"

	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/storage"
	pb "github.com/Leo-MathGuy/YandexLMS_Final/internal/grpc"
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

func (s *taskServer) RecieveTask(ctx context.Context, req *pb.TaskSubmit) (*pb.Empty, error) {
	return &pb.Empty{}, storage.FinishTask(&storage.T, uint(req.Id), req.Result)
}

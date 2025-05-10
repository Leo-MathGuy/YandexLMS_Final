package agent

import (
	"context"
	"time"

	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/logging"
	pb "github.com/Leo-MathGuy/YandexLMS_Final/internal/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	workerCount = 5
	adress      = ":5050"
)

func createWorker(ctx context.Context, client pb.TasksClient, id int) {
	logging.Log("Agent thread %d started", id)
	for {
		select {
		case <-ctx.Done():
			logging.Log("Agent thread %d exiting", id)
		default:
			resp, err := client.GetTask(ctx, &pb.Empty{})

			if err != nil {
				logging.Error("Error getting task from %d: %s", id, err.Error())
				continue
			}

			if !resp.Have {
				time.Sleep(1 * time.Second)
				return
			}

			result := func() float64 {
				defer func() {
					if r := recover(); r != nil {
						logging.Error("Recovered from reading response: %s", r)
					}
				}()

				time.Sleep(time.Duration(resp.OpTime) * time.Millisecond)

				switch resp.Operator {
				case "+":
					return resp.Left + resp.Right
				case "-":
					return resp.Left - resp.Right
				case "*":
					return resp.Left * resp.Right
				case "/":
					return resp.Left / resp.Right
				}

				panic("No op")
			}()

			for i := range 5 {
				_, err = client.SubmitTask(ctx, &pb.TaskSubmit{Id: resp.Id, Result: result})

				if err == nil {
					break
				}
				logging.Error("Agent thread %d could not send task [%d/5]: %s", id, i, err.Error())
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func StartThreads(ctx context.Context) {
	var conn *grpc.ClientConn
	for i := range 5 {
		var err error
		conn, err = grpc.NewClient(adress, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			continue
		}
		if i == 4 && err != nil {
			logging.Panic("Failed to connect")
		}
	}
	defer conn.Close()
	client := pb.NewTasksClient(conn)

	for v := range 5 {
		go createWorker(ctx, client, v)
	}

	logging.Log("Agent threads created")
}

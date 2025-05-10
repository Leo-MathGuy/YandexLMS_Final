package agent

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Leo-MathGuy/YandexLMS_Final/internal/app/logging"
	pb "github.com/Leo-MathGuy/YandexLMS_Final/internal/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	workerCount = 1
	address     = ":5050"
)

func createWorker(ctx context.Context, client pb.TasksClient, id int) {
	disconnected := false

	logging.Log("Agent thread %d started", id)
	for {
		select {
		case <-ctx.Done():
			logging.Log("Agent thread %d exiting", id)
			return
		default:
			resp, err := client.GetTask(ctx, &pb.Empty{})

			if err != nil {
				if strings.Contains(err.Error(), "dialing") {
					if !disconnected {
						logging.Error("Agent %d cannot connect to server", id)
						disconnected = true
					}
				} else if !strings.Contains(err.Error(), "context canceled") {
					logging.Error("Error getting task in %d: %s", id, err.Error())
				}
				time.Sleep(
					time.Second +
						time.Duration(rand.Int31n(1000))*time.Millisecond,
				)
				// Stagger errors
				continue
			}

			if disconnected {
				logging.Log("Agent %d reconnected", id)
				disconnected = false
			}

			if !resp.Have {
				time.Sleep(1 * time.Second)
				continue
			}

			result, err := func() (float64, error) {
				defer func() {
					if r := recover(); r != nil {
						logging.Error("Recovered from reading response: %s", r)
					}
				}()

				time.Sleep(time.Duration(resp.OpTime) * time.Millisecond)

				switch resp.Operator {
				case "+":
					return resp.Left + resp.Right, nil
				case "-":
					return resp.Left - resp.Right, nil
				case "*":
					return resp.Left * resp.Right, nil
				case "/":
					if resp.Right == 0 {
						// Division by 0 is a good way for a malicious
						// Attacker to DoS the server for a while
						// Because of a lack of pre-checks but
						// I do not want to add yet another check
						// So just assume it would be used by good people
						// :)
						return 0, fmt.Errorf("division by 0")
					}
					return resp.Left / resp.Right, nil
				}

				panic("No op")
			}()

			for i := range 5 {
				er := ""
				if err == nil {
					er = ""
				} else {
					er = err.Error()
				}
				_, err = client.SubmitTask(ctx, &pb.TaskSubmit{Id: resp.Id, Result: result, Error: er})

				if err == nil {
					break
				}
				logging.Error("Agent thread %d could not send task [%d/5]: %s", id, i, err.Error())
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func StartThreads(ctx context.Context) *grpc.ClientConn {
	var conn *grpc.ClientConn
	for i := range 5 {
		var err error
		conn, err = grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			continue
		}
		if i == 4 {
			logging.Panic("Failed to connect")
		}
	}
	client := pb.NewTasksClient(conn)

	for v := range 5 {
		go createWorker(ctx, client, v)
	}

	logging.Log("Agent threads created")
	return conn
}

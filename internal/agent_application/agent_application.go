package agentapplication

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	pb "github.com/ERRORIK404/Distribution_gRPC_Calculator/pkg/proto"
	structs "github.com/ERRORIK404/Distribution_gRPC_Calculator/pkg/structs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)
type Agent struct {
  Agent http.Client
}

func New() *Agent {
  return &Agent{http.Client{}}
}

func AgentRun() error{
	// url := "http://localhost:8080/internal/task"
	ch := make(chan struct{})
	go func() {
		for {
			// a := New()
			time.Sleep(time.Second)
			// response := new(http.Response)
			// response, err := a.Agent.Get(url)
            // if response.StatusCode != http.StatusOK ||  err != nil{
            //     log.Println("Unexpected status code:", response.StatusCode)
            //     continue
            // }
			conn, err := grpc.Dial("localhost:8081", grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Fatalf("did not connect: %v", err)
				conn.Close()
				continue
			}
			c := pb.NewOrchestratorServiceClient(conn)
			in := &pb.EmptyMessage{}
			res, err := c.GetTask(context.Background(), in)
			if err != nil {
				conn.Close()
				continue
			}
			task := structs.Task{Id: res.Task.Id, Operand1: float64(res.Task.Operand1), Operand2: float64(res.Task.Operand2), Operation: res.Task.Operation, Operation_time: int(res.Task.OperationTime)}

			timer := time.NewTimer(time.Duration(task.Operation_time) * time.Millisecond)	
			var result structs.Result
			result.Id = task.Id
			switch task.Operation {
				case "+":
					result.Result = task.Operand1 + task.Operand2
				case "-":
					result.Result = task.Operand1 - task.Operand2
				case "*":
					result.Result = task.Operand1 * task.Operand2
				case "/":
					result.Result = task.Operand1 / task.Operand2
			}
			// jsonData, _ := json.Marshal(result)
			<- timer.C
			// http.Post(url, "application/json", bytes.NewBuffer(jsonData))
			sendres := &pb.ResultRequest{Result: &pb.Result{Id: result.Id, Result: float32(result.Result)}}
			c.SendResult(context.TODO(), sendres)
			conn.Close()
		}
	}()
	<-ch
	defer log.Println("DIED")
	return errors.New("agent died")
}
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/xyzbit/minitaskx/core/model"
)

const (
	host = "http://localhost:8080/"
)

var rootCmd = &cobra.Command{
	Use:   "minitaskx-example",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application.`,
}

var createTaskCmd = &cobra.Command{
	Use:   "create",
	Short: "Run a new task",
	RunE: func(cmd *cobra.Command, args []string) error {
		bizID := fmt.Sprintf("test_flow_run_go_%d", rand.Int31())
		if err := createTask(bizID); err != nil {
			return err
		}
		fmt.Println("create task success, will be finshed after 30s, bizID: ", bizID)
		return nil
	},
}

var createDockerTaskCmd = &cobra.Command{
	Use:   "create-docker",
	Short: "Run a new docker task",
	RunE: func(cmd *cobra.Command, args []string) error {
		bizID := fmt.Sprintf("test_flow_run_docker_%d", rand.Int31())
		if err := createDockerTask(bizID); err != nil {
			return err
		}
		fmt.Println("create task success, will be finshed after 30s, bizID: ", bizID)
		return nil
	},
}

var createK8sJobTaskCmd = &cobra.Command{
	Use:   "create-k8sjob",
	Short: "Run a new k8sjob task",
	RunE: func(cmd *cobra.Command, args []string) error {
		bizID := fmt.Sprintf("test_flow_run_k8sjob_%d", rand.Int31())
		if err := createK8sJobTask(bizID); err != nil {
			return err
		}
		fmt.Println("create task success, will be finshed after 30s, bizID: ", bizID)
		return nil
	},
}

var (
	createCount    int
	createTaskNCmd = &cobra.Command{
		Use:   "createn",
		Short: "Run many new task",
		RunE: func(cmd *cobra.Command, args []string) error {
			for i := 0; i < createCount; i++ {
				bizID := fmt.Sprintf("test_flow_run_%d", rand.Int31())
				if err := createTask(bizID); err != nil {
					return err
				}
				fmt.Printf("create task %d success, will be finished after 30s, bizID: %s\n", i+1, bizID)
			}
			return nil
		},
	}
)

var watchTaskCmd = &cobra.Command{
	Use:   "watch",
	Short: "watch task status",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("please input bizID")
		}
		bizID := args[0]
		return watchTaskStatus(bizID)
	},
}

var pauseTaskCmd = &cobra.Command{
	Use:   "pause",
	Short: "pause task",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("please input bizID")
		}
		bizID := args[0]
		return operateTask(bizID, model.TaskStatusPaused)
	},
}

var resumeTaskCmd = &cobra.Command{
	Use:   "resume",
	Short: "resume task",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("please input bizID")
		}
		bizID := args[0]
		return operateTask(bizID, model.TaskStatusRunning)
	},
}

var stopTaskCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop task",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("please input bizID")
		}
		bizID := args[0]
		return operateTask(bizID, model.TaskStatusStop)
	},
}

func init() {
	rootCmd.AddCommand(createTaskCmd)
	createTaskNCmd.Flags().IntVarP(&createCount, "count", "n", 1, "number of tasks to create")
	rootCmd.AddCommand(createTaskNCmd)
	rootCmd.AddCommand(createDockerTaskCmd)
	rootCmd.AddCommand(createK8sJobTaskCmd)
	rootCmd.AddCommand(pauseTaskCmd)
	rootCmd.AddCommand(resumeTaskCmd)
	rootCmd.AddCommand(stopTaskCmd)
	rootCmd.AddCommand(watchTaskCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

func createTask(bizID string) error {
	var req bytes.Buffer

	d := fmt.Sprintf(`
		{
			"biz_id": "%s",
			"biz_type": "test",
			"type": "goroutine",
			"payload": "test task run"
		}
	`, bizID)
	req.WriteString(d)
	return sendCreateRequest(&req)
}

func createDockerTask(bizID string) error {
	payload := `
	{
		"image": "busybox:latest",
		"cmd": ["sh", "-c", "for i in $(seq 1 30); do echo \"hello $i\"; sleep 1; done"]
	}
	`

	fmt.Printf("payload: %s\n bizID: %s", payload, bizID)
	task := &model.Task{
		BizID:   bizID,
		BizType: "test",
		Type:    "docker",
		Payload: payload,
	}
	req, _ := json.Marshal(task)
	return sendCreateRequest(bytes.NewBuffer(req))
}

func createK8sJobTask(bizID string) error {
	payload := fmt.Sprintf(`
	{
	    "name": "%s",
		"image": "busybox:latest",
		"command": ["sh", "-c", "for i in $(seq 1 30); do echo \"hello $i\"; sleep 1; done"]
	}`, bizID)

	task := &model.Task{
		BizID:   bizID,
		BizType: "test",
		Type:    "k8sjob",
		Payload: payload,
	}
	req, _ := json.Marshal(task)
	return sendCreateRequest(bytes.NewBuffer(req))
}

func sendCreateRequest(body io.Reader) error {
	resp, err := http.Post(host+"api/v1/scheduler/tasks/create", "application/json", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if status := resp.StatusCode; status != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("got %d want %d, resp: %s", status, http.StatusOK, string(body))
	}

	return nil
}

func operateTask(bizID string, status model.TaskStatus) error {
	req := bytes.NewBufferString(fmt.Sprintf(`
		{
			"biz_id": "%s",
			"status": "%s"
		}
	`, bizID, status.String()))
	resp, err := http.Post(host+"api/v1/scheduler/tasks/operate", "application/json", req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if status := resp.StatusCode; status != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("got %d want %d, resp: %s", status, http.StatusOK, string(body))
	}
	return nil
}

func watchTaskStatus(bizID string) error {
	for {
		time.Sleep(time.Second)

		task, err := listTasks(bizID)
		if err != nil {
			return err
		}

		if task == nil {
			continue
		}
		fmt.Printf("task status: %s\n", task.Status)
	}
}

func listTasks(bizID string) (*model.Task, error) {
	resp, err := http.Get(host + "api/v1/scheduler/tasks/list?biz_ids=" + bizID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data struct {
		Data []*model.Task `json:"data"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	if len(data.Data) == 0 {
		return nil, nil
	}
	return data.Data[0], nil
}

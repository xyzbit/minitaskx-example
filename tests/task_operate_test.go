package tests

import (
	"bytes"
	"fmt"
	"net/http"
)

func operateTask(bizID string) error {
	var req bytes.Buffer
	d := fmt.Sprintf(`
		{
			"biz_id": "%s",
			"type": "goroutine",
			"payload": "test task run"
		}
	`, bizID)
	req.WriteString(d)
	resp, err := http.Post("api/v1/scheduler/tasks/operate", "application/json", &req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if status := resp.StatusCode; status != http.StatusOK {
		return fmt.Errorf("got %d want %d", status, http.StatusOK)
	}

	return nil
}

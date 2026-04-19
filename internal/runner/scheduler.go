package runner

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

type taskSpec struct {
	Name    string
	Timeout time.Duration
	Soft    bool
	Run     func(context.Context) error
}

func runTasks(ctx context.Context, tasks ...taskSpec) []string {
	var wg sync.WaitGroup
	errCh := make(chan string, len(tasks))

	for _, task := range tasks {
		if task.Run == nil {
			continue
		}
		wg.Add(1)
		go func(task taskSpec) {
			defer wg.Done()

			taskCtx := ctx
			cancel := func() {}
			if task.Timeout > 0 {
				taskCtx, cancel = context.WithTimeout(ctx, task.Timeout)
			}
			defer cancel()

			if err := task.Run(taskCtx); err != nil {
				if taskCtx.Err() == context.DeadlineExceeded {
					errCh <- fmt.Sprintf("采集任务 %s 超时，已降级跳过: %v", task.Name, taskCtx.Err())
					return
				}
				if task.Soft {
					errCh <- fmt.Sprintf("采集任务 %s 已降级: %v", task.Name, err)
					return
				}
				errCh <- fmt.Sprintf("采集任务 %s 失败: %v", task.Name, err)
			}
		}(task)
	}

	wg.Wait()
	close(errCh)

	errs := make([]string, 0, len(tasks))
	for err := range errCh {
		errs = append(errs, err)
	}
	sort.Strings(errs)
	return errs
}

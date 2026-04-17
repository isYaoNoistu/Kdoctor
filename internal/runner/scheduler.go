package runner

import "sync"

func parallel(tasks ...func()) {
	var wg sync.WaitGroup
	for _, task := range tasks {
		if task == nil {
			continue
		}
		wg.Add(1)
		go func(task func()) {
			defer wg.Done()
			task()
		}(task)
	}
	wg.Wait()
}

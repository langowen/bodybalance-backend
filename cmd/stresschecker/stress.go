package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

var (
	numWorkers        = 5    // Количество параллельных горутин
	requestsPerWorker = 1000 // Сколько запросов отсылает одна горутина
)

type Request struct {
	URL string
}

func main() {
	var wg sync.WaitGroup

	requestList := []Request{
		{"https://api.7375.org/v1/video?type=basic&category=Шея"},
		{"https://api.7375.org/v1/category?type=basic"},
		{"https://api.7375.org/v1/login?username=base&type=Basic"},
	}

	fmt.Printf("Количество работников: %d\n", numWorkers)
	fmt.Printf("Запросов на каждого работника: %d\n", requestsPerWorker)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			runWorker(requestList, workerID, requestsPerWorker)
		}(i + 1)
	}

	wg.Wait()
	fmt.Println("Все рабочие закончили выполнение.")
}

// Запуск одной рабочей горутины
func runWorker(requests []Request, workerID int, count int) {
	startTime := time.Now()
	totalRequestsSent := 0
	totalResponseTime := time.Duration(0)

	for j := 0; j < count; j++ {
		reqIndex := j % len(requests)
		url := requests[reqIndex].URL
		responseTime := sendGetRequest(url)
		if responseTime != 0 {
			totalResponseTime += responseTime
			totalRequestsSent++
		}
	}

	endTime := time.Since(startTime)
	avgResponseTime := float64(totalResponseTime.Nanoseconds()) / float64(time.Millisecond) / float64(totalRequestsSent)

	fmt.Printf("Рабочий #%d завершил работу:\n", workerID)
	fmt.Printf("\tОтправлено запросов: %d\n", totalRequestsSent)
	fmt.Printf("\tОбщее время обработки: %.2f секунд\n", endTime.Seconds())
	fmt.Printf("\tСреднее время ответа сервера: %.2f миллисекунд\n", avgResponseTime)
}

// Отправляет один GET-запрос и возвращает время ответа
func sendGetRequest(url string) time.Duration {
	start := time.Now()
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode >= 400 {
		return 0
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	return time.Since(start)
}

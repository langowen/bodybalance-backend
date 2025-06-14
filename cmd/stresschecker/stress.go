package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	numWorkers        = 5                      // Количество параллельных горутин
	requestsPerWorker = 1000                   // Сколько запросов отсылает одна горутина
	baseURL           = "https://api.7375.org" // Базовый URL по умолчанию
	requestSetID      = 2                      // ID набора запросов по умолчанию
	durationMinutes   = 0                      // Продолжительность теста в минутах (0 = использовать фиксированное количество запросов)
	byTime            = false                  // Флаг для выбора режима работы (по времени или по количеству запросов)
	maxConnsPerHost   = 100                    // Максимальное количество соединений на один хост
	maxIdleConns      = 100                    // Максимальное количество простаивающих соединений
	timeoutSeconds    = 30                     // Таймаут запроса в секундах
	skipResponseBody  = true                   // Не читать тело ответа для ускорения
	disableKeepAlives = false                  // Отключить постоянные соединения
)

// Path представляет путь запроса без базового URL
type Path struct {
	Path string
}

// Request представляет полный URL запроса
type Request struct {
	URL string
}

// WorkerResult Результаты работы воркера
type WorkerResult struct {
	RequestsSent    int
	TotalTime       time.Duration
	AvgResponseTime float64
}

func main() {
	var wg sync.WaitGroup
	resultsCh := make(chan WorkerResult, 100) // Канал для сбора результатов

	requestSets := map[int][]Path{
		1: { // Видео и API запросы
			{"/video/Vstuplenie.mp4"},
			{"/video/Koleno_baza.mp4"},
			{"/video/sample-5s.mp4"},
			{"/v1/video_categories?type=2&category=1"},
			{"/v1/video_categories?type=3&category=3"},
			{"/v1/category?type=1"},
			{"/v1/category?type=3"},
			{"/v1/login?username=base"},
		},
		2: { // Только API запросы
			{"/v1/video?video_id=3"},
			{"/v1/video?video_id=4"},
			{"/v1/video?video_id=1"},
			{"/v1/category?type=2"},
			{"/v1/login?username=base"},
			{"/v1/video_categories?type=1&category=1"},
			{"/v1/video_categories?type=1&category=3"},
		},
		3: { // Только видеофайлы
			{"/video/Vstuplenie.mp4"},
			{"/video/Koleno_baza.mp4"},
			{"/video/sample-5s.mp4"},
			{"/video/Golenostop_baza.mp4"},
			{"/video/Plecho_baza.mp4"},
		},
	}

	// Настройка флагов командной строки
	flag.IntVar(&requestsPerWorker, "requests", requestsPerWorker, "Количество запросов на одного воркера")
	flag.IntVar(&numWorkers, "workers", numWorkers, "Количество параллельных воркеров")
	flag.StringVar(&baseURL, "base", baseURL, "Базовый URL (например, https://body.7375.org или http://localhost:8083)")
	flag.IntVar(&requestSetID, "set", requestSetID, "ID набора запросов (1: смешанный, 2: только API, 3: только видео)")
	flag.IntVar(&durationMinutes, "duration", durationMinutes, "Продолжительность теста в минутах (0 = фиксированное количество запросов)")
	flag.BoolVar(&byTime, "bytime", byTime, "Режим работы: true = по времени, false = по количеству запросов")
	flag.IntVar(&maxConnsPerHost, "maxconns", maxConnsPerHost, "Максимальное количество соединений на хост")
	flag.IntVar(&maxIdleConns, "maxidle", maxIdleConns, "Максимальное количество простаивающих соединений")
	flag.IntVar(&timeoutSeconds, "timeout", timeoutSeconds, "Таймаут запроса в секундах")
	flag.BoolVar(&skipResponseBody, "skipbody", skipResponseBody, "Пропускать чтение тела ответа для ускорения")
	flag.BoolVar(&disableKeepAlives, "nokeepalive", disableKeepAlives, "Отключить постоянные соединения")
	flag.Parse()

	paths, ok := requestSets[requestSetID]
	if !ok {
		fmt.Printf("Ошибка: набор запросов с ID %d не существует\n", requestSetID)
		return
	}

	requestList := make([]Request, len(paths))
	for i, path := range paths {
		requestList[i] = Request{baseURL + path.Path}
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(timeoutSeconds) * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          maxIdleConns,
		MaxIdleConnsPerHost:   maxConnsPerHost,
		MaxConnsPerHost:       maxConnsPerHost,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableKeepAlives:     disableKeepAlives,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(timeoutSeconds) * time.Second,
	}

	fmt.Println("=========== Настройки теста ===========")
	fmt.Printf("Базовый URL: %s\n", baseURL)
	fmt.Printf("Используемый набор запросов: %d\n", requestSetID)
	fmt.Printf("Количество работников: %d\n", numWorkers)
	if byTime {
		fmt.Printf("Режим: по времени (%d минут)\n", durationMinutes)
	} else {
		fmt.Printf("Режим: по количеству запросов (%d на каждого работника)\n", requestsPerWorker)
	}
	fmt.Println("======== Настройки оптимизации ========")
	fmt.Printf("Максимум соединений на хост: %d\n", maxConnsPerHost)
	fmt.Printf("Максимум простаивающих соединений: %d\n", maxIdleConns)
	fmt.Printf("Таймаут запроса: %d секунд\n", timeoutSeconds)
	fmt.Printf("Пропуск чтения тела ответа: %v\n", skipResponseBody)
	fmt.Printf("Отключить постоянные соединения: %v\n", disableKeepAlives)
	fmt.Println("=======================================")

	startTime := time.Now()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			result := runWorker(requestList, workerID, requestsPerWorker, durationMinutes, byTime, httpClient)
			resultsCh <- result
		}(i + 1)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var totalResults WorkerResult
	for result := range resultsCh {
		totalResults.RequestsSent += result.RequestsSent
		if result.TotalTime > totalResults.TotalTime {
			totalResults.TotalTime = result.TotalTime
		}
	}

	totalTime := time.Since(startTime)
	reqPerSecond := float64(totalResults.RequestsSent) / totalTime.Seconds()

	fmt.Println("\n============ ИТОГИ ТЕСТА ============")
	fmt.Printf("Общее время выполнения: %.2f секунд\n", totalTime.Seconds())
	fmt.Printf("Всего отправлено запросов: %d\n", totalResults.RequestsSent)
	fmt.Printf("Запросов в секунду: %.2f\n", reqPerSecond)
	fmt.Println("=======================================")
}

// Запуск одной рабочей горутины
func runWorker(requests []Request, workerID int, count int, duration int, byTime bool, client *http.Client) WorkerResult {
	startTime := time.Now()
	totalRequestsSent := 0
	totalResponseTime := time.Duration(0)
	result := WorkerResult{}

	for {
		batchSize := 100 // Размер пакета запросов
		if !byTime && count < batchSize {
			batchSize = count
		}

		for j := 0; j < batchSize; j++ {
			reqIndex := (totalRequestsSent) % len(requests)
			url := requests[reqIndex].URL
			responseTime := sendGetRequest(url, client)
			if responseTime != 0 {
				totalResponseTime += responseTime
				totalRequestsSent++
			}

			if !byTime && totalRequestsSent >= count {
				break
			}
		}

		if byTime {
			if time.Since(startTime).Minutes() >= float64(duration) {
				break
			}
		} else {
			if totalRequestsSent >= count {
				break
			}
		}
	}

	endTime := time.Since(startTime)
	avgResponseTime := float64(0)
	if totalRequestsSent > 0 {
		avgResponseTime = float64(totalResponseTime.Nanoseconds()) / float64(time.Millisecond) / float64(totalRequestsSent)
	}

	fmt.Printf("[Воркер #%d] Отправлено: %d запросов за %.2f сек. (%.2f запросов/сек, среднее время ответа: %.2f мс)\n",
		workerID, totalRequestsSent, endTime.Seconds(),
		float64(totalRequestsSent)/endTime.Seconds(), avgResponseTime)

	result.RequestsSent = totalRequestsSent
	result.TotalTime = endTime
	result.AvgResponseTime = avgResponseTime

	return result
}

// Функция отправки запроса
func sendGetRequest(url string, client *http.Client) time.Duration {
	start := time.Now()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0
	}

	// Добавляем заголовок OP
	req.Header.Add("OP", "stresscheck")

	resp, err := client.Do(req)
	if err != nil {
		return 0
	}

	if resp.StatusCode >= 400 {
		resp.Body.Close()
		return 0
	}

	// Опционально пропускаем чтение тела ответа для ускорения
	if skipResponseBody {
		resp.Body.Close()
	} else {
		// Если нужно прочитать тело, читаем и закрываем
		_, err := io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		if err != nil {
			return 0
		}
	}

	return time.Since(start)
}

package api

import (
	"bytes"
	"context"
	"fmt"
	"go-course/global"
	"go-course/model"
	"go-course/mq"
	redisrepo "go-course/repository/redis"
	"go-course/utils"
	"io"
	"math"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type benchmarkRequest struct {
	CourseID uint   `json:"course_id"`
	Stock    int    `json:"stock"`
	Users    int    `json:"users"`
	Duration string `json:"duration"`
}

type benchmarkPoint struct {
	Label string `json:"label"`
	P50   int64  `json:"p50"`
	P90   int64  `json:"p90"`
	P99   int64  `json:"p99"`
	QPS   int    `json:"qps"`
}

type benchmarkSnapshot struct {
	Running      bool             `json:"running"`
	Finished     bool             `json:"finished"`
	Countdown    int              `json:"countdown"`
	Elapsed      int              `json:"elapsed"`
	TotalSeconds int              `json:"total_seconds"`
	Metrics      benchmarkMetrics `json:"metrics"`
	Monitor      benchmarkMonitor `json:"monitor"`
	Points       []benchmarkPoint `json:"points"`
	Message      string           `json:"message"`
}

type benchmarkMetrics struct {
	QPS          int    `json:"qps"`
	AvgLatency   int64  `json:"avg_latency"`
	P99Latency   int64  `json:"p99_latency"`
	Success      int64  `json:"success"`
	Failed       int64  `json:"failed"`
	OversoldText string `json:"oversold_text"`
}

type benchmarkMonitor struct {
	RedisStock  int    `json:"redis_stock"`
	Queued      int    `json:"queued"`
	Processing  int    `json:"processing"`
	DLQ         int    `json:"dlq"`
	Written     int64  `json:"written"`
	MQPublished uint64 `json:"mq_published"`
	MQConsumed  uint64 `json:"mq_consumed"`
	MQBacklog   int    `json:"mq_backlog"`
}

type benchmarkBucket struct {
	sync.Mutex
	latencies []int64
	total     int
	success   int64
	failed    int64
}

var benchmarkRunner = struct {
	sync.Mutex
	cancel context.CancelFunc
	state  benchmarkSnapshot
}{
	state: benchmarkSnapshot{
		Metrics: benchmarkMetrics{OversoldText: "—"},
		Message: "等待压测开始",
	},
}

func StartBenchmark(c *gin.Context) {
	var req benchmarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": "压测参数错误"})
		return
	}
	if req.CourseID == 0 {
		req.CourseID = 1
	}
	if req.Stock <= 0 {
		req.Stock = 1000
	}
	if req.Users < 1 {
		req.Users = 10
	}
	if req.Users > 500 {
		req.Users = 500
	}
	seconds := parseBenchmarkDuration(req.Duration)

	benchmarkRunner.Lock()
	if benchmarkRunner.state.Running {
		benchmarkRunner.Unlock()
		c.JSON(http.StatusConflict, gin.H{"code": http.StatusConflict, "msg": "压测正在进行中"})
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	benchmarkRunner.cancel = cancel
	benchmarkRunner.state = benchmarkSnapshot{
		Running:      true,
		Finished:     false,
		Countdown:    seconds,
		Elapsed:      0,
		TotalSeconds: seconds,
		Metrics:      benchmarkMetrics{OversoldText: "验证中"},
		Monitor:      benchmarkMonitor{RedisStock: req.Stock},
		Points:       []benchmarkPoint{},
		Message:      "真实压测已启动",
	}
	snapshot := cloneBenchmarkSnapshotLocked()
	benchmarkRunner.Unlock()

	go runBenchmark(ctx, req, seconds)
	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "真实压测已启动", "data": snapshot})
}

func GetBenchmarkStatus(c *gin.Context) {
	benchmarkRunner.Lock()
	snapshot := cloneBenchmarkSnapshotLocked()
	benchmarkRunner.Unlock()
	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "ok", "data": snapshot})
}

func runBenchmark(ctx context.Context, req benchmarkRequest, seconds int) {
	if err := resetBenchmarkData(req.CourseID, req.Stock); err != nil {
		global.Logger.Error("压测环境重置失败", zap.Error(err))
		finishBenchmarkWithError("压测环境重置失败")
		return
	}

	bucket := &benchmarkBucket{}
	runCtx, cancel := context.WithTimeout(ctx, time.Duration(seconds)*time.Second)
	defer cancel()

	client := &http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        req.Users * 2,
			MaxIdleConnsPerHost: req.Users * 2,
			IdleConnTimeout:     30 * time.Second,
		},
	}

	var seq uint64
	var wg sync.WaitGroup
	baseStudentID := uint64(time.Now().Unix()%100000) * 1000000
	for i := 0; i < req.Users; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-runCtx.Done():
					return
				default:
				}
				next := atomic.AddUint64(&seq, 1)
				latency, ok := sendBenchmarkRequest(runCtx, client, req.CourseID, uint(baseStudentID+next))
				bucket.add(latency, ok)
			}
		}()
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	totalSuccess := int64(0)
	totalFailed := int64(0)
	points := make([]benchmarkPoint, 0, seconds)
	startedAt := time.Now()

	for elapsed := 1; elapsed <= seconds; elapsed++ {
		select {
		case <-runCtx.Done():
			elapsed = seconds
		case <-ticker.C:
		}

		tick := bucket.drain()
		totalSuccess += tick.success
		totalFailed += tick.failed
		point := benchmarkPoint{
			Label: fmt.Sprintf("%ds", elapsed),
			P50:   percentile(tick.latencies, 50),
			P90:   percentile(tick.latencies, 90),
			P99:   percentile(tick.latencies, 99),
			QPS:   tick.total,
		}
		points = append(points, point)
		if len(points) > 60 {
			points = points[len(points)-60:]
		}
		monitor := loadBenchmarkMonitor(req.CourseID)
		updateBenchmarkState(benchmarkSnapshot{
			Running:      elapsed < seconds,
			Finished:     elapsed >= seconds,
			Countdown:    maxInt(seconds-elapsed, 0),
			Elapsed:      elapsed,
			TotalSeconds: seconds,
			Metrics: benchmarkMetrics{
				QPS:          tick.total,
				AvgLatency:   avgLatency(tick.latencies),
				P99Latency:   point.P99,
				Success:      totalSuccess,
				Failed:       totalFailed,
				OversoldText: oversoldText(elapsed < seconds, monitor.Written, req.Stock),
			},
			Monitor: monitor,
			Points:  append([]benchmarkPoint(nil), points...),
			Message: "真实压测进行中",
		})
		if elapsed >= seconds || time.Since(startedAt) >= time.Duration(seconds)*time.Second {
			break
		}
	}

	cancel()
	wg.Wait()
	monitor := loadBenchmarkMonitor(req.CourseID)
	benchmarkRunner.Lock()
	benchmarkRunner.state.Running = false
	benchmarkRunner.state.Finished = true
	benchmarkRunner.state.Countdown = 0
	benchmarkRunner.state.Monitor = monitor
	benchmarkRunner.state.Metrics.OversoldText = oversoldText(false, monitor.Written, req.Stock)
	benchmarkRunner.state.Message = "真实压测已结束"
	benchmarkRunner.cancel = nil
	benchmarkRunner.Unlock()
}

func sendBenchmarkRequest(ctx context.Context, client *http.Client, courseID, studentID uint) (int64, bool) {
	token, err := utils.GenToken(studentID, fmt.Sprintf("bench-%d", studentID))
	if err != nil {
		return 0, false
	}
	body := bytes.NewReader(nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://127.0.0.1:8080/auth/select/%d", courseID), body)
	if err != nil {
		return 0, false
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return latency, false
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return latency, resp.StatusCode >= 200 && resp.StatusCode < 300
}

func resetBenchmarkData(courseID uint, stock int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := global.DB.WithContext(ctx).Model(&model.Course{}).Where("id = ?", courseID).Update("stock", stock).Error; err != nil {
		return err
	}
	if err := global.DB.WithContext(ctx).Where("course_id = ?", courseID).Delete(&model.Selection{}).Error; err != nil {
		return err
	}
	stockKey := fmt.Sprintf("course:stock:%d", courseID)
	if err := global.RDB.Set(ctx, stockKey, stock, 0).Err(); err != nil {
		return err
	}
	_ = global.RDB.Del(ctx, fmt.Sprintf("course:%d", courseID)).Err()
	_ = global.RDB.XTrimMaxLen(ctx, redisrepo.SelectStreamKey, 0).Err()
	mq.ResetMetrics()
	purgeQueue(global.Settings.RabbitMQ.QueueName)
	purgeQueue(global.Settings.RabbitMQ.QueueName + ".retry.1s")
	purgeQueue(global.Settings.RabbitMQ.QueueName + ".retry.5s")
	purgeQueue(global.Settings.RabbitMQ.QueueName + ".retry.10s")
	purgeQueue(global.Settings.RabbitMQ.QueueName + ".dlq")
	return nil
}

func purgeQueue(name string) {
	if global.MQChannel == nil {
		return
	}
	if _, err := global.MQChannel.QueuePurge(name, false); err != nil {
		global.Logger.Warn("压测前清空队列失败", zap.String("queue", name), zap.Error(err))
	}
}

func loadBenchmarkMonitor(courseID uint) benchmarkMonitor {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	redisStock := 0
	if value, err := global.RDB.Get(ctx, fmt.Sprintf("course:stock:%d", courseID)).Int(); err == nil {
		redisStock = value
	}
	var written int64
	_ = global.DB.WithContext(ctx).Model(&model.Selection{}).
		Where("course_id = ? AND status = ?", courseID, model.SelectionStatusSelected).
		Count(&written).Error
	mqSnapshot := mq.SnapshotMetrics()
	backlog := queueMessages(global.Settings.RabbitMQ.QueueName)
	return benchmarkMonitor{
		RedisStock:  redisStock,
		Queued:      int(mqSnapshot.Published),
		Processing:  int(mqSnapshot.Consumed),
		DLQ:         int(mqSnapshot.DLQ),
		Written:     written,
		MQPublished: mqSnapshot.Published,
		MQConsumed:  mqSnapshot.Consumed,
		MQBacklog:   backlog,
	}
}

func queueMessages(name string) int {
	if global.MQChannel == nil {
		return 0
	}
	queue, err := global.MQChannel.QueueInspect(name)
	if err != nil {
		return 0
	}
	return queue.Messages
}

func (b *benchmarkBucket) add(latency int64, success bool) {
	b.Lock()
	b.latencies = append(b.latencies, latency)
	b.total++
	if success {
		b.success++
	} else {
		b.failed++
	}
	b.Unlock()
}

func (b *benchmarkBucket) drain() benchmarkBucket {
	b.Lock()
	defer b.Unlock()
	next := benchmarkBucket{
		latencies: append([]int64(nil), b.latencies...),
		total:     b.total,
		success:   b.success,
		failed:    b.failed,
	}
	b.latencies = b.latencies[:0]
	b.total = 0
	b.success = 0
	b.failed = 0
	return next
}

func updateBenchmarkState(snapshot benchmarkSnapshot) {
	benchmarkRunner.Lock()
	benchmarkRunner.state = snapshot
	benchmarkRunner.Unlock()
}

func finishBenchmarkWithError(message string) {
	benchmarkRunner.Lock()
	benchmarkRunner.state.Running = false
	benchmarkRunner.state.Finished = true
	benchmarkRunner.state.Message = message
	benchmarkRunner.state.Metrics.OversoldText = "异常"
	benchmarkRunner.cancel = nil
	benchmarkRunner.Unlock()
}

func cloneBenchmarkSnapshotLocked() benchmarkSnapshot {
	snapshot := benchmarkRunner.state
	snapshot.Points = append([]benchmarkPoint(nil), benchmarkRunner.state.Points...)
	return snapshot
}

func parseBenchmarkDuration(value string) int {
	switch value {
	case "10s":
		return 10
	case "60s":
		return 60
	default:
		return 30
	}
}

func percentile(values []int64, p int) int64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]int64(nil), values...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	index := int(math.Ceil(float64(p)/100*float64(len(sorted)))) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return sorted[index]
}

func avgLatency(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	total := int64(0)
	for _, value := range values {
		total += value
	}
	return total / int64(len(values))
}

func oversoldText(running bool, written int64, stock int) string {
	if running {
		return "验证中"
	}
	if written <= int64(stock) {
		return "通过"
	}
	return "异常"
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

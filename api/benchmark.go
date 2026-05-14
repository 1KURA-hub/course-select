package api

import (
	"bytes"
	"context"
	"encoding/json"
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
	"strings"
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
	Limits       benchmarkLimits  `json:"limits"`
}

type benchmarkMetrics struct {
	QPS          int               `json:"qps"`
	AvgLatency   int64             `json:"avg_latency"`
	P99Latency   int64             `json:"p99_latency"`
	Success      int64             `json:"success"`
	Failed       int64             `json:"failed"`
	OversoldText string            `json:"oversold_text"`
	Failures     benchmarkFailures `json:"failures"`
}

type benchmarkFailures struct {
	Unauthorized int64 `json:"unauthorized"`
	StockEmpty   int64 `json:"stock_empty"`
	Duplicate    int64 `json:"duplicate"`
	ServerError  int64 `json:"server_error"`
	NetworkError int64 `json:"network_error"`
	Other        int64 `json:"other"`
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
	failures  benchmarkFailures
}

type benchmarkLimits struct {
	MaxStock             int  `json:"max_stock"`
	MaxUsers             int  `json:"max_users"`
	MaxSeconds           int  `json:"max_seconds"`
	LargeStockThreshold  int  `json:"large_stock_threshold"`
	LargeStockCooldown   int  `json:"large_stock_cooldown"`
	SecondsUntilNextRun  int  `json:"seconds_until_next_run"`
	LargeStockRestricted bool `json:"large_stock_restricted"`
}

var benchmarkRunner = struct {
	sync.Mutex
	cancel       context.CancelFunc
	state        benchmarkSnapshot
	lastLargeRun time.Time
}{
	state: benchmarkSnapshot{
		Metrics: benchmarkMetrics{OversoldText: "—"},
		Message: "等待压测开始",
		Limits:  currentBenchmarkLimits(0),
	},
}

const (
	benchmarkMaxStock            = 5000
	benchmarkMaxUsers            = 200
	benchmarkMaxSeconds          = 60
	benchmarkLargeStockThreshold = 1000
	benchmarkLargeStockCooldown  = 3 * time.Minute
)

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
	if req.Stock > benchmarkMaxStock {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": fmt.Sprintf("课程总库存不能超过 %d", benchmarkMaxStock)})
		return
	}
	if req.Users < 1 {
		req.Users = 10
	}
	if req.Users > benchmarkMaxUsers {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": fmt.Sprintf("并发用户数不能超过 %d", benchmarkMaxUsers)})
		return
	}
	seconds := parseBenchmarkDuration(req.Duration)
	if seconds > benchmarkMaxSeconds {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": fmt.Sprintf("压测时长不能超过 %d 秒", benchmarkMaxSeconds)})
		return
	}

	benchmarkRunner.Lock()
	if benchmarkRunner.state.Running {
		benchmarkRunner.Unlock()
		c.JSON(http.StatusConflict, gin.H{"code": http.StatusConflict, "msg": "压测正在进行中"})
		return
	}
	if req.Stock > benchmarkLargeStockThreshold {
		nextAllowedAt := benchmarkRunner.lastLargeRun.Add(benchmarkLargeStockCooldown)
		if !benchmarkRunner.lastLargeRun.IsZero() && time.Now().Before(nextAllowedAt) {
			waitSeconds := int(math.Ceil(time.Until(nextAllowedAt).Seconds()))
			benchmarkRunner.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code": http.StatusTooManyRequests,
				"msg":  fmt.Sprintf("高库存压测冷却中，请 %d 秒后重试", waitSeconds),
				"data": benchmarkSnapshot{Limits: currentBenchmarkLimits(waitSeconds)},
			})
			return
		}
		benchmarkRunner.lastLargeRun = time.Now()
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
		Limits:       currentBenchmarkLimits(0),
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
				latency, ok, reason := sendBenchmarkRequest(runCtx, client, req.CourseID, uint(baseStudentID+next))
				bucket.add(latency, ok, reason)
			}
		}()
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	totalSuccess := int64(0)
	totalFailed := int64(0)
	totalFailures := benchmarkFailures{}
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
		totalFailures.add(tick.failures)
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
				Failures:     totalFailures,
			},
			Monitor: monitor,
			Points:  append([]benchmarkPoint(nil), points...),
			Message: "真实压测进行中",
			Limits:  currentBenchmarkLimits(0),
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
	benchmarkRunner.state.Limits = currentBenchmarkLimits(0)
	benchmarkRunner.cancel = nil
	benchmarkRunner.Unlock()
}

func sendBenchmarkRequest(ctx context.Context, client *http.Client, courseID, studentID uint) (int64, bool, string) {
	token, err := utils.GenToken(studentID, fmt.Sprintf("bench-%d", studentID))
	if err != nil {
		return 0, false, "server_error"
	}
	body := bytes.NewReader(nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://127.0.0.1:8080/auth/select/%d", courseID), body)
	if err != nil {
		return 0, false, "server_error"
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return latency, false, "network_error"
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return latency, true, ""
	}
	return latency, false, classifyBenchmarkFailure(resp.StatusCode, raw)
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

func (b *benchmarkBucket) add(latency int64, success bool, reason string) {
	b.Lock()
	b.latencies = append(b.latencies, latency)
	b.total++
	if success {
		b.success++
	} else {
		b.failed++
		b.failures.inc(reason)
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
		failures:  b.failures,
	}
	b.latencies = b.latencies[:0]
	b.total = 0
	b.success = 0
	b.failed = 0
	b.failures = benchmarkFailures{}
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
	benchmarkRunner.state.Limits = currentBenchmarkLimits(0)
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

func currentBenchmarkLimits(secondsUntilNextRun int) benchmarkLimits {
	return benchmarkLimits{
		MaxStock:             benchmarkMaxStock,
		MaxUsers:             benchmarkMaxUsers,
		MaxSeconds:           benchmarkMaxSeconds,
		LargeStockThreshold:  benchmarkLargeStockThreshold,
		LargeStockCooldown:   int(benchmarkLargeStockCooldown.Seconds()),
		SecondsUntilNextRun:  maxInt(secondsUntilNextRun, 0),
		LargeStockRestricted: secondsUntilNextRun > 0,
	}
}

func classifyBenchmarkFailure(statusCode int, body []byte) string {
	if statusCode == http.StatusUnauthorized {
		return "unauthorized"
	}
	msg := string(body)
	var payload struct {
		Msg string `json:"msg"`
	}
	if err := json.Unmarshal(body, &payload); err == nil && payload.Msg != "" {
		msg = payload.Msg
	}
	switch {
	case strings.Contains(msg, "库存不足"):
		return "stock_empty"
	case strings.Contains(msg, "重复") || strings.Contains(msg, "不可重复"):
		return "duplicate"
	case statusCode >= http.StatusInternalServerError || strings.Contains(msg, "系统繁忙"):
		return "server_error"
	default:
		return "other"
	}
}

func (f *benchmarkFailures) inc(reason string) {
	switch reason {
	case "unauthorized":
		f.Unauthorized++
	case "stock_empty":
		f.StockEmpty++
	case "duplicate":
		f.Duplicate++
	case "server_error":
		f.ServerError++
	case "network_error":
		f.NetworkError++
	default:
		f.Other++
	}
}

func (f *benchmarkFailures) add(next benchmarkFailures) {
	f.Unauthorized += next.Unauthorized
	f.StockEmpty += next.StockEmpty
	f.Duplicate += next.Duplicate
	f.ServerError += next.ServerError
	f.NetworkError += next.NetworkError
	f.Other += next.Other
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

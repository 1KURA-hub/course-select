import { useEffect, useRef, useState } from "react";
import { Loader2, Play } from "lucide-react";
import { api } from "../api";
import { MetricCard } from "../components/MetricCard";
import type { BenchmarkFailures, BenchmarkPoint, BenchmarkStatus } from "../types";

type Metric = { label: string; value: string; tone: string };
type Monitor = { redisStock: number; queued: number; processing: number; dlq: number; written: number; mqPublished: number; mqConsumed: number; mqBacklog: number };

const maxStock = 5000;
const maxUsers = 200;
const largeStockThreshold = 1000;
const emptyFailures: BenchmarkFailures = { unauthorized: 0, stock_empty: 0, duplicate: 0, server_error: 0, network_error: 0, other: 0 };

const zeroMetrics: Metric[] = [
  { label: "QPS", value: "0", tone: "qps" },
  { label: "平均延迟", value: "0ms", tone: "latency" },
  { label: "P99 延迟", value: "0ms", tone: "p99" },
  { label: "成功请求", value: "0", tone: "success" },
  { label: "失败请求", value: "0", tone: "failed" },
  { label: "不超卖验证", value: "—", tone: "safe" }
];

const initialMonitor: Monitor = { redisStock: 1000, queued: 0, processing: 0, dlq: 0, written: 0, mqPublished: 0, mqConsumed: 0, mqBacklog: 0 };

export function PerformancePage() {
  const [stock, setStock] = useState(1000);
  const [users, setUsers] = useState(120);
  const [duration, setDuration] = useState("30s");
  const [running, setRunning] = useState(false);
  const [countdown, setCountdown] = useState(0);
  const [elapsed, setElapsed] = useState(0);
  const [totalSeconds, setTotalSeconds] = useState(30);
  const [chartPoints, setChartPoints] = useState<BenchmarkPoint[]>([]);
  const [metrics, setMetrics] = useState<Metric[]>(zeroMetrics);
  const [monitor, setMonitor] = useState<Monitor>(initialMonitor);
  const [failures, setFailures] = useState<BenchmarkFailures>(emptyFailures);
  const [finished, setFinished] = useState(false);
  const [notice, setNotice] = useState("等待真实压测开始");
  const [error, setError] = useState("");

  useEffect(() => {
    if (!running) return;
    let active = true;

    async function syncStatus() {
      try {
        const payload = await api.getBenchmarkStatus();
        if (!active || !payload.data) return;
        applyStatus(payload.data);
      } catch (err) {
        if (!active) return;
        setError(err instanceof Error ? err.message : "获取压测状态失败");
      }
    }

    void syncStatus();
    const timer = window.setInterval(syncStatus, 1000);
    return () => {
      active = false;
      window.clearInterval(timer);
    };
  }, [running]);

  async function startBenchmark() {
    setError("");
    setFinished(false);
    setRunning(true);
    setCountdown(Number(duration.replace("s", "")));
    setElapsed(0);
    setChartPoints([]);
    setMetrics([{ ...zeroMetrics[0] }, { ...zeroMetrics[1] }, { ...zeroMetrics[2] }, { ...zeroMetrics[3] }, { ...zeroMetrics[4] }, { label: "不超卖验证", value: "验证中", tone: "safe" }]);
    setMonitor({ ...initialMonitor, redisStock: normalizedStock });
    setFailures(emptyFailures);
    try {
      const payload = await api.startBenchmark({ stock: normalizedStock, users: normalizedUsers, duration, course_id: 1 });
      if (payload.data) applyStatus(payload.data);
    } catch (err) {
      setRunning(false);
      setError(err instanceof Error ? err.message : "启动真实压测失败");
    }
  }

  const normalizedStock = Math.min(Math.max(stock || 1, 1), maxStock);
  const normalizedUsers = Math.min(Math.max(users || 1, 1), maxUsers);

  function applyStatus(status: BenchmarkStatus) {
    setRunning(status.running);
    setFinished(status.finished);
    setCountdown(status.countdown);
    setElapsed(status.elapsed);
    setTotalSeconds(status.total_seconds);
    setNotice(status.message || (status.running ? "真实压测进行中" : "真实压测已结束"));
    setMetrics(metricsFromStatus(status));
    setMonitor({
      redisStock: status.monitor.redis_stock,
      queued: status.monitor.queued,
      processing: status.monitor.processing,
      dlq: status.monitor.dlq,
      written: status.monitor.written,
      mqPublished: status.monitor.mq_published ?? status.monitor.queued,
      mqConsumed: status.monitor.mq_consumed ?? status.monitor.processing,
      mqBacklog: status.monitor.mq_backlog ?? status.monitor.queued
    });
    setFailures(status.metrics.failures || emptyFailures);
    setChartPoints(status.points || []);
  }

  return (
    <section className="page performance-page">
      <div className="page-heading">
        <span className="eyebrow">Performance Board</span>
        <h1>性能看板</h1>
        <p>点击开始压测后，服务器会真实并发请求选课接口，完整经过 JWT、Redis Lua、Redis Stream、RabbitMQ 与 MySQL 链路。</p>
      </div>

      <div className="benchmark-panel">
        <label>
          课程总库存
          <input
            type="number"
            min={1}
            max={maxStock}
            value={stock}
            disabled={running}
            onBlur={() => setStock(normalizedStock)}
            onChange={(event) => setStock(Number(event.target.value))}
          />
        </label>
        <label>
          并发用户数
          <input type="range" min={10} max={maxUsers} value={users} disabled={running} onChange={(event) => setUsers(Number(event.target.value))} />
          <span>{users}</span>
        </label>
        <label>
          持续时间
          <select value={duration} disabled={running} onChange={(event) => setDuration(event.target.value)}>
            <option>10s</option>
            <option>30s</option>
            <option>60s</option>
          </select>
        </label>
        <button className="primary-button" disabled={running || normalizedStock !== stock || normalizedUsers !== users} onClick={startBenchmark}>
          {running ? <Loader2 className="spin" size={16} /> : <Play size={16} />}
          {running ? `压测中... ${countdown}s` : "开始真实压测"}
        </button>
      </div>

      <div className="benchmark-notice">
        <span>{notice} · 库存≤1000不限次数，库存&gt;1000 每3分钟一次</span>
        <b>{elapsed > 0 ? `${elapsed}/${totalSeconds}s` : "READY"}</b>
      </div>
      {stock > largeStockThreshold ? <div className="benchmark-guard">高库存真实压测已开启保护：当前库存超过 1000，后端会限制 3 分钟内只能启动一次。</div> : null}
      {normalizedStock !== stock || normalizedUsers !== users ? <div className="form-notice">参数超出安全范围：库存最高 {maxStock}，并发最高 {maxUsers}。</div> : null}
      {error ? <div className="form-notice">{error}</div> : null}

      <div className="metric-grid">
        {metrics.map((item) => <MetricCard key={item.label} {...item} />)}
      </div>

      <div className="realtime-panel">
        <div className="realtime-card">
          <strong>Redis 库存</strong>
          <small>剩余库存</small>
          <span key={monitor.redisStock}>{monitor.redisStock}</span>
          <div className="stock-drain">
            <i
              className={monitor.redisStock / Math.max(normalizedStock, 1) < 0.1 ? "danger" : monitor.redisStock / Math.max(normalizedStock, 1) <= 0.3 ? "warn" : ""}
              style={{ width: `${Math.max(0, (monitor.redisStock / Math.max(normalizedStock, 1)) * 100)}%` }}
            />
          </div>
        </div>
        <div className="realtime-card">
          <strong>RabbitMQ 队列</strong>
          <div className="queue-mini">
            <span>已投递 <b key={monitor.mqPublished}>{monitor.mqPublished}</b></span>
            <span>已消费 <b key={monitor.mqConsumed}>{monitor.mqConsumed}</b></span>
            <span>积压 <b key={monitor.mqBacklog}>{monitor.mqBacklog}</b></span>
          </div>
        </div>
        <div className="realtime-card">
          <strong>MySQL 落库</strong>
          <span key={monitor.written}>{monitor.written}</span>
          <MiniLineCanvas written={monitor.written} stock={normalizedStock} />
        </div>
      </div>

      <div className="failure-panel">
        <strong>失败原因诊断</strong>
        <span>JWT 鉴权 <b>{failures.unauthorized}</b></span>
        <span>库存不足 <b>{failures.stock_empty}</b></span>
        <span>重复选课 <b>{failures.duplicate}</b></span>
        <span>服务错误 <b>{failures.server_error}</b></span>
        <span>网络错误 <b>{failures.network_error}</b></span>
        <span>其他 <b>{failures.other}</b></span>
      </div>

      <div className="chart-panel">
        <LatencyChart points={chartPoints} />
        <ThroughputChart points={chartPoints} finished={finished} />
      </div>
    </section>
  );
}

function metricsFromStatus(status: BenchmarkStatus): Metric[] {
  return [
    { label: "QPS", value: status.metrics.qps.toLocaleString(), tone: "qps" },
    { label: "平均延迟", value: `${status.metrics.avg_latency}ms`, tone: "latency" },
    { label: "P99 延迟", value: `${status.metrics.p99_latency}ms`, tone: "p99" },
    { label: "成功请求", value: status.metrics.success.toLocaleString(), tone: "success" },
    { label: "失败请求", value: status.metrics.failed.toLocaleString(), tone: "failed" },
    { label: "不超卖验证", value: status.metrics.oversold_text || "—", tone: "safe" }
  ];
}

function MiniLineCanvas({ written, stock }: { written: number; stock: number }) {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;
    const ratio = window.devicePixelRatio || 1;
    const width = 240;
    const height = 60;
    canvas.width = width * ratio;
    canvas.height = height * ratio;
    canvas.style.width = "100%";
    canvas.style.height = `${height}px`;
    ctx.setTransform(ratio, 0, 0, ratio, 0, 0);
    ctx.clearRect(0, 0, width, height);
    const end = Math.min(written / Math.max(stock, 1), 1);
    const values = [0.04, 0.12, 0.2, 0.34, 0.48, 0.62, 0.78, end];
    ctx.beginPath();
    values.forEach((value, index) => {
      const x = (index / (values.length - 1)) * width;
      const y = height - 8 - value * 44;
      if (index === 0) ctx.moveTo(x, y);
      else ctx.lineTo(x, y);
    });
    ctx.strokeStyle = "#10b981";
    ctx.lineWidth = 1.5;
    ctx.lineCap = "round";
    ctx.lineJoin = "round";
    ctx.stroke();
  }, [stock, written]);

  return <canvas className="mini-line" ref={canvasRef} aria-label="MySQL 落库增长趋势" />;
}

function LatencyChart({ points }: { points: BenchmarkPoint[] }) {
  return (
    <div className="line-chart-card">
      <strong>延迟分布</strong>
      <div className="chart-legend">
        <span className="p50">P50</span><span className="p90">P90</span><span className="p99">P99</span>
      </div>
      <svg viewBox="0 0 100 46" preserveAspectRatio="none">
        <polyline className="p50" points={fixedLinePoints(points.map((item) => item.p50), 110)} />
        <polyline className="p90" points={fixedLinePoints(points.map((item) => item.p90), 110)} />
        <polyline className="p99" points={fixedLinePoints(points.map((item) => item.p99), 110)} />
        {tickMarks(points).map((tick) => (
          <text className="chart-tick" x={tick.x} y="45" textAnchor="middle" key={tick.label}>{tick.label}</text>
        ))}
      </svg>
    </div>
  );
}

function ThroughputChart({ points, finished }: { points: BenchmarkPoint[]; finished: boolean }) {
  return (
    <div className="line-chart-card">
      <strong>吞吐量 QPS</strong>
      <div className="chart-legend"><span className="qps">QPS</span></div>
      <svg viewBox="0 0 100 46" preserveAspectRatio="none">
        {points.length === 0 ? <text x="50" y="24" textAnchor="middle" className="chart-empty">等待压测开始</text> : null}
        <polygon className="qps-fill" points={fixedAreaPoints(points.map((item) => item.qps), 8000)} />
        <polyline className="qps" points={fixedLinePoints(points.map((item) => item.qps), 8000)} />
        {tickMarks(points).map((tick) => (
          <text className="chart-tick" x={tick.x} y="45" textAnchor="middle" key={tick.label}>{tick.label}</text>
        ))}
        {finished && points.length > 0 ? (
          <>
            <line className="end-marker" x1="100" y1="4" x2="100" y2="42" />
            <text x="96" y="8" textAnchor="end" className="end-label">压测结束</text>
          </>
        ) : null}
      </svg>
    </div>
  );
}

function fixedLinePoints(values: number[], max: number) {
  if (values.length === 0) return "";
  if (values.length === 1) return `0,${42 - Math.min(values[0] / max, 1) * 34}`;
  return values.map((value, index) => `${(index / (values.length - 1)) * 100},${42 - Math.min(value / max, 1) * 34}`).join(" ");
}

function fixedAreaPoints(values: number[], max: number) {
  if (values.length === 0) return "";
  return `0,46 ${fixedLinePoints(values, max)} 100,46`;
}

function tickMarks(points: BenchmarkPoint[]) {
  if (points.length <= 1) return [];
  return points
    .map((point, index) => ({ label: point.label, x: (index / (points.length - 1)) * 100, index }))
    .filter((tick) => tick.index === 0 || tick.index === points.length - 1 || tick.index % 5 === 4);
}

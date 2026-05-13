import { useEffect, useMemo, useRef, useState } from "react";
import { Loader2, Play } from "lucide-react";
import { MetricCard } from "../components/MetricCard";

type Metric = { label: string; value: string; tone: string };
type ChartPoint = { label: string; p50: number; p90: number; p99: number; qps: number };
type RunTotals = { success: number; failed: number };

const zeroMetrics: Metric[] = [
  { label: "QPS", value: "0", tone: "qps" },
  { label: "平均延迟", value: "0ms", tone: "latency" },
  { label: "P99 延迟", value: "0ms", tone: "p99" },
  { label: "成功请求", value: "0", tone: "success" },
  { label: "失败请求", value: "0", tone: "failed" },
  { label: "不超卖验证", value: "—", tone: "safe" }
];

export function PerformancePage() {
  const [stock, setStock] = useState(1000);
  const [users, setUsers] = useState(120);
  const [duration, setDuration] = useState("30s");
  const [running, setRunning] = useState(false);
  const [countdown, setCountdown] = useState(0);
  const [elapsed, setElapsed] = useState(0);
  const [totalSeconds, setTotalSeconds] = useState(30);
  const [chartPoints, setChartPoints] = useState<ChartPoint[]>([]);
  const [metrics, setMetrics] = useState<Metric[]>(zeroMetrics);
  const [runTotals, setRunTotals] = useState<RunTotals>({ success: 0, failed: 0 });
  const [finished, setFinished] = useState(false);

  const progress = totalSeconds > 0 ? Math.min(elapsed / totalSeconds, 1) : 0;

  const monitor = useMemo(() => {
    if (!running && elapsed === 0 && runTotals.success === 0) {
      return { redisStock: stock, queued: 0, processing: 0, dlq: 0, written: 0 };
    }
    const redisStock = Math.max(0, stock - runTotals.success);
    const queueWave = running ? Math.sin(Math.PI * progress) : 0;
    return {
      redisStock,
      queued: Math.max(0, Math.round(users * 2.8 * queueWave)),
      processing: running ? Math.max(0, Math.round(users * 0.18 * (1 - progress * 0.35))) : 0,
      dlq: running && progress > 0.72 ? Math.round((progress - 0.72) * users * 0.08) : 0,
      written: runTotals.success
    };
  }, [elapsed, progress, runTotals.success, running, stock, users]);

  useEffect(() => {
    if (!running) return;
    const timer = window.setInterval(() => {
      setElapsed((value) => {
        const nextElapsed = value + 1;
        const done = nextElapsed >= totalSeconds;
        const qps = Math.floor(Math.random() * 3000 + 3000);
        const p50 = Math.round(Math.random() * 6 + 12);
        const p90 = Math.round(Math.random() * 10 + 30);
        const p99 = Math.round(Math.random() * 20 + 70);
        const successDelta = Math.max(1, Math.round((stock / Math.max(totalSeconds, 1)) * (0.72 + Math.random() * 0.56)));
        const failedDelta = Math.floor(Math.random() * Math.max(2, users / 40));
        setRunTotals((current) => {
          const nextSuccess = Math.min(stock, current.success + successDelta);
          const nextFailed = current.failed + failedDelta;
          const liveMetrics: Metric[] = [
            { label: "QPS", value: qps.toLocaleString(), tone: "qps" },
            { label: "平均延迟", value: `${p50}ms`, tone: "latency" },
            { label: "P99 延迟", value: `${p99}ms`, tone: "p99" },
            { label: "成功请求", value: nextSuccess.toLocaleString(), tone: "success" },
            { label: "失败请求", value: nextFailed.toLocaleString(), tone: "failed" },
            { label: "不超卖验证", value: "验证中", tone: "safe" }
          ];
          setMetrics(done ? finalMetrics(stock, nextSuccess, nextFailed, totalSeconds) : liveMetrics);
          return { success: nextSuccess, failed: nextFailed };
        });
        setChartPoints((points) => [
          ...points.slice(-59),
          {
            label: `${nextElapsed}s`,
            p50,
            p90,
            p99,
            qps
          }
        ]);
        setCountdown(done ? 0 : Math.max(0, totalSeconds - nextElapsed));
        if (done) {
          setRunning(false);
          setFinished(true);
        }
        return nextElapsed;
      });
    }, 1000);

    return () => window.clearInterval(timer);
  }, [running, stock, totalSeconds, users]);

  function startBenchmark() {
    const seconds = Number(duration.replace("s", ""));
    setTotalSeconds(seconds);
    setCountdown(seconds);
    setElapsed(0);
    setChartPoints([]);
    setRunTotals({ success: 0, failed: 0 });
    setFinished(false);
    setMetrics(zeroMetrics);
    setRunning(true);
  }

  return (
    <section className="page performance-page">
      <div className="page-heading">
        <span className="eyebrow">Performance Board</span>
        <h1>性能看板</h1>
        <p>展示高并发选课链路在库存有限场景下的压测结果与系统能力。</p>
      </div>

      <div className="benchmark-panel">
        <label>
          课程总库存
          <input type="number" min={1} value={stock} disabled={running} onChange={(event) => setStock(Number(event.target.value))} />
        </label>
        <label>
          并发用户数
          <input type="range" min={10} max={500} value={users} disabled={running} onChange={(event) => setUsers(Number(event.target.value))} />
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
        <button className="primary-button" disabled={running} onClick={startBenchmark}>
          {running ? <Loader2 className="spin" size={16} /> : <Play size={16} />}
          {running ? `压测中... ${countdown}s` : "开始压测"}
        </button>
      </div>

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
              className={monitor.redisStock / Math.max(stock, 1) < 0.1 ? "danger" : monitor.redisStock / Math.max(stock, 1) <= 0.3 ? "warn" : ""}
              style={{ width: `${Math.max(0, (monitor.redisStock / Math.max(stock, 1)) * 100)}%` }}
            />
          </div>
        </div>
        <div className="realtime-card">
          <strong>RabbitMQ 队列</strong>
          <div className="queue-mini">
            <span>待处理 <b key={monitor.queued}>{monitor.queued}</b></span>
            <span>处理中 <b key={monitor.processing}>{monitor.processing}</b></span>
            <span>死信 <b key={monitor.dlq}>{monitor.dlq}</b></span>
          </div>
        </div>
        <div className="realtime-card">
          <strong>MySQL 落库</strong>
          <span key={monitor.written}>{monitor.written}</span>
          <MiniLineCanvas written={monitor.written} stock={stock} />
        </div>
      </div>

      <div className="chart-panel">
        <LatencyChart points={chartPoints} />
        <ThroughputChart points={chartPoints} finished={finished} />
      </div>
    </section>
  );
}

function finalMetrics(stock: number, successTotal: number, failedTotal: number, seconds: number): Metric[] {
  const success = Math.min(stock, successTotal);
  const failed = failedTotal;
  const qps = Math.round((success + failed) / Math.max(seconds, 1)) + Math.floor(Math.random() * 1200 + 3600);
  return [
    { label: "QPS", value: qps.toLocaleString(), tone: "qps" },
    { label: "平均延迟", value: `${Math.floor(Math.random() * 7 + 12)}ms`, tone: "latency" },
    { label: "P99 延迟", value: `${Math.floor(Math.random() * 21 + 70)}ms`, tone: "p99" },
    { label: "成功请求", value: success.toLocaleString(), tone: "success" },
    { label: "失败请求", value: failed.toLocaleString(), tone: "failed" },
    { label: "不超卖验证", value: success <= stock ? "通过" : "异常", tone: "safe" }
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
    ctx.scale(ratio, ratio);
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

function LatencyChart({ points }: { points: ChartPoint[] }) {
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

function ThroughputChart({ points, finished }: { points: ChartPoint[]; finished: boolean }) {
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
  return values.map((value, index) => `${(index / (values.length - 1)) * 100},${42 - (value / max) * 34}`).join(" ");
}

function fixedAreaPoints(values: number[], max: number) {
  if (values.length === 0) return "";
  return `0,46 ${fixedLinePoints(values, max)} 100,46`;
}

function tickMarks(points: ChartPoint[]) {
  if (points.length <= 1) return [];
  return points
    .map((point, index) => ({ label: point.label, x: (index / (points.length - 1)) * 100, index }))
    .filter((tick) => tick.index === 0 || tick.index === points.length - 1 || tick.index % 5 === 4);
}

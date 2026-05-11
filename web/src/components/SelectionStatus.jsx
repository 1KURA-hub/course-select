import { useState, useEffect, useRef } from "react";
import { CheckCircle2, XCircle, Loader2 } from "lucide-react";
import { api } from "../api";

const STEPS = [
  { key: "submit", label: "请求提交" },
  { key: "deduct", label: "库存预扣" },
  { key: "queue", label: "排队中" },
  { key: "done", label: "处理完成" },
];

const particles = Array.from({ length: 10 }, (_, i) => ({
  tx: (Math.random() - 0.5) * 80 + "px",
  ty: (Math.random() - 0.5) * 80 + "px",
  delay: Math.random() * 0.3 + "s",
}));

export default function SelectionStatus({ courseId, onDone }) {
  const [status, setStatus] = useState("polling");
  const [stepIndex, setStepIndex] = useState(0);
  const [shakeKey, setShakeKey] = useState(0);
  const timerRef = useRef(null);

  useEffect(() => {
    let step = 0;
    const stepTimer = setInterval(() => { step++; setStepIndex(Math.min(step, 3)); }, 600);
    timerRef.current = setInterval(async () => {
      try {
        const data = await api.getResult(courseId);
        if (data.status === "success") {
          setStatus("success");
          setStepIndex(3);
          clearStepTimers();
        } else if (data.status === "failed") {
          setStatus("failed");
          setStepIndex(3);
          setShakeKey((k) => k + 1);
          clearStepTimers();
        }
      } catch { /* keep polling */ }
    }, 1500);

    function clearStepTimers() {
      clearInterval(stepTimer);
      if (timerRef.current) clearInterval(timerRef.current);
    }

    return clearStepTimers;
  }, [courseId]);

  return (
    <div className="status-modal" onClick={(e) => e.target === e.currentTarget && onDone()}>
      <div className={`glass status-card ${status === "failed" ? "shake" : ""}`} key={shakeKey}>
        {status === "success" && (
          <div className="particles">
            {particles.map((p, i) => (
              <div key={i} className="particle" style={{ "--tx": p.tx, "--ty": p.ty, animationDelay: p.delay, left: "50%", top: "50%" }} />
            ))}
            <CheckCircle2 size={48} color="var(--success)" style={{ position: "absolute", inset: 0, margin: "auto" }} />
          </div>
        )}
        {status === "failed" && (
          <div className="icon-wrap">
            <XCircle size={48} color="var(--danger)" />
          </div>
        )}
        {status === "polling" && (
          <div className="icon-wrap">
            <Loader2 size={48} color="var(--primary)" className="spin" style={{ animation: "spin 1s linear infinite" }} />
          </div>
        )}

        <div className="status-steps">
          {STEPS.map((s, i) => {
            const done = i < stepIndex || status !== "polling";
            const active = i === stepIndex && status === "polling";
            const err = status === "failed" && i === 3;
            return (
              <div key={s.key} style={{ display: "flex", alignItems: "center" }}>
                <div className="status-step">
                  <div className={`dot ${done ? "done" : ""} ${active ? "active" : ""} ${err ? "done" : ""}`}
                    style={err ? { background: "var(--danger)", borderColor: "var(--danger)" } : {}} />
                  <span className={`label ${active || done ? "on" : ""}`}>{s.label}</span>
                </div>
                {i < 3 && <div className={`status-line ${done ? "done" : ""}`} />}
              </div>
            );
          })}
        </div>

        {status === "polling" && (
          <>
            <p className="status-text" style={{ color: "var(--primary)" }}>选课处理中</p>
            <p className="status-sub">您的请求已提交，正在排队等待系统处理...</p>
          </>
        )}
        {status === "success" && (
          <>
            <p className="status-text" style={{ color: "var(--success)" }}>选课成功</p>
            <p className="status-sub">库存扣减完成，选课记录已保存</p>
            <button className="btn btn-primary" style={{ marginTop: 8 }} onClick={onDone}>确定</button>
          </>
        )}
        {status === "failed" && (
          <>
            <p className="status-text" style={{ color: "var(--danger)" }}>选课失败</p>
            <p className="status-sub">库存不足或系统繁忙，请稍后重试</p>
            <button className="btn btn-ghost" style={{ marginTop: 8 }} onClick={onDone}>关闭</button>
          </>
        )}
      </div>
    </div>
  );
}

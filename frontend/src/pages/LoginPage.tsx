import { FormEvent } from "react";
import { ArrowRight, Loader2, Zap } from "lucide-react";

export function LoginPage({
  sid,
  password,
  loading,
  notice,
  onSid,
  onPassword,
  onSubmit
}: {
  sid: string;
  password: string;
  loading: boolean;
  notice: string;
  onSid: (value: string) => void;
  onPassword: (value: string) => void;
  onSubmit: (event: FormEvent) => void;
}) {
  return (
    <section className="login-page">
      <div className="login-brand-panel">
        <span className="eyebrow"><Zap size={14} /> CourseRush</span>
        <h1>CourseRush 高并发选课系统</h1>
        <p>毫秒级抢课 · Redis 预扣库存 · RabbitMQ 异步削峰</p>
        <div className="login-feature-grid">
          <span>Redis Lua 原子扣减</span>
          <span>Stream Outbox 防丢</span>
          <span>MySQL 最终一致</span>
          <span>不超卖压测验证</span>
        </div>
      </div>

      <form className="login-card" onSubmit={onSubmit}>
        <span className="eyebrow">Student Access</span>
        <h2>登录选课控制台</h2>
        <p>进入后可以查看课程库存、提交抢课请求并追踪异步处理链路。</p>
        <label>
          用户 ID
          <input value={sid} onChange={(event) => onSid(event.target.value)} placeholder="请输入学号" />
        </label>
        <label>
          密码
          <input type="password" value={password} onChange={(event) => onPassword(event.target.value)} placeholder="请输入密码" />
        </label>
        {notice ? <div className="form-notice">{notice}</div> : null}
        <button className="primary-button login-submit" disabled={loading || !sid || !password}>
          {loading ? <Loader2 className="spin" size={18} /> : <ArrowRight size={18} />}
          登录
        </button>
      </form>
    </section>
  );
}

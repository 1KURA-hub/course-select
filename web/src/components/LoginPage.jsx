import { useState } from "react";
import { Sparkles } from "lucide-react";
import { api } from "../api";

export default function LoginPage({ onLogin }) {
  const [isRegister, setIsRegister] = useState(false);
  const [sid, setSid] = useState("");
  const [password, setPassword] = useState("");
  const [name, setName] = useState("");
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e) {
    e.preventDefault();
    setErr("");
    setLoading(true);
    try {
      let data;
      if (isRegister) {
        await api.register(sid, password, name || sid);
        data = await api.login(sid, password);
      } else {
        data = await api.login(sid, password);
      }
      localStorage.setItem("token", data.token || "");
      onLogin({ sid, name: data.name || sid });
    } catch (e) {
      setErr(e.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="login-bg">
      <div className="glass login-card">
        <div className="logo"><Sparkles size={28} color="#fff" /></div>
        <h1>{isRegister ? "创建账号" : "欢迎回来"}</h1>
        <p className="sub">
          {isRegister ? "加入高并发选课系统" : "登录高并发选课系统"}
        </p>
        <form onSubmit={handleSubmit}>
          <div className="input-wrap">
            <input id="sid" value={sid} onChange={(e) => setSid(e.target.value)} placeholder=" " />
            <label htmlFor="sid">学号</label>
          </div>
          <div className="input-wrap">
            <input id="pw" type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder=" " />
            <label htmlFor="pw">密码</label>
          </div>
          {isRegister && (
            <div className="input-wrap">
              <input id="nm" value={name} onChange={(e) => setName(e.target.value)} placeholder=" " />
              <label htmlFor="nm">姓名（选填）</label>
            </div>
          )}
          {err && <p style={{ color: "var(--danger)", fontSize: 13, marginBottom: 8 }}>{err}</p>}
          <button className="btn btn-primary" type="submit" disabled={loading || !sid || !password}>
            {loading ? "处理中..." : isRegister ? "注册" : "登录"}
          </button>
        </form>
        <p className="switch">
          {isRegister ? "已有账号？" : "没有账号？"}
          <button onClick={() => { setIsRegister(!isRegister); setErr(""); }}>
            {isRegister ? "去登录" : "去注册"}
          </button>
        </p>
      </div>
    </div>
  );
}

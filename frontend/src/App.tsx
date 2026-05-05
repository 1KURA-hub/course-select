import {
  Activity,
  ArrowRight,
  BookOpen,
  CheckCircle2,
  Clock3,
  Database,
  Loader2,
  LogOut,
  MessageSquare,
  Rabbit,
  RotateCcw,
  Search,
  Server,
  ShieldCheck,
  Sparkles,
  UserRound,
  XCircle
} from "lucide-react";
import { FormEvent, useCallback, useEffect, useMemo, useRef, useState } from "react";

type Course = {
  ID: number;
  Name: string;
  Stock: number;
  TeacherID: number;
};

type Selection = {
  selection_id: number;
  student_id: number;
  course_id: number;
  status: number;
  status_text: string;
  course_name: string;
  teacher_id: number;
};

type ApiResponse<T = unknown> = {
  code?: number;
  msg?: string;
  data?: T;
  token?: string;
  name?: string;
  id?: number;
  course_id?: number;
};

type AuthUser = {
  token: string;
  name: string;
  id: number;
};

type ResultState = "idle" | "pending" | "success" | "failed" | "dropped";
type AuthMode = "login" | "register";

const timeline = [
  { title: "发起请求", desc: "Gin 鉴权与参数校验", icon: Server },
  { title: "Lua 预扣", desc: "请求去重与库存扣减", icon: ShieldCheck },
  { title: "Stream 写入", desc: "保存预扣成功消息", icon: MessageSquare },
  { title: "RabbitMQ", desc: "异步削峰消费", icon: Rabbit },
  { title: "MySQL 落库", desc: "联合唯一索引兜底", icon: Database },
  { title: "结果返回", desc: "轮询 request 状态", icon: CheckCircle2 }
];

const tokenKey = "course_select_token";
const userKey = "course_select_user";

function loadAuth(): AuthUser | null {
  const token = localStorage.getItem(tokenKey);
  const raw = localStorage.getItem(userKey);
  if (!token || !raw) return null;
  try {
    const user = JSON.parse(raw) as Omit<AuthUser, "token">;
    return { ...user, token };
  } catch {
    return null;
  }
}

async function apiRequest<T>(path: string, token?: string, options: RequestInit = {}) {
  const headers = new Headers(options.headers);
  if (token) headers.set("Authorization", `Bearer ${token}`);
  if (options.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const resp = await fetch(path, { ...options, headers });
  const text = await resp.text();
  const payload = text ? (JSON.parse(text) as ApiResponse<T>) : {};

  if (!resp.ok) {
    throw new Error(payload.msg || `请求失败: ${resp.status}`);
  }
  return payload;
}

function normalizeCourses(data: unknown): Course[] {
  if (!Array.isArray(data)) return [];
  return data.map((item) => item as Course);
}

function resultFromMessage(message?: string): ResultState {
  if (!message) return "pending";
  if (message.includes("成功")) return "success";
  if (message.includes("失败")) return "failed";
  if (message.includes("已退课")) return "dropped";
  return "pending";
}

function statusLabel(state: ResultState) {
  switch (state) {
    case "pending":
      return "排队中";
    case "success":
      return "抢课成功";
    case "failed":
      return "抢课失败";
    case "dropped":
      return "已退课";
    default:
      return "等待操作";
  }
}

export function App() {
  const [auth, setAuth] = useState<AuthUser | null>(() => loadAuth());
  const [authMode, setAuthMode] = useState<AuthMode>("login");
  const [sid, setSid] = useState("");
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const [courses, setCourses] = useState<Course[]>([]);
  const [selectedCourse, setSelectedCourse] = useState<Course | null>(null);
  const [selections, setSelections] = useState<Selection[]>([]);
  const [query, setQuery] = useState("");
  const [resultState, setResultState] = useState<ResultState>("idle");
  const [activeStep, setActiveStep] = useState(0);
  const [loading, setLoading] = useState(false);
  const [busyAction, setBusyAction] = useState<"select" | "drop" | null>(null);
  const [notice, setNotice] = useState("登录后加载课程列表，选择课程后可以开始演示选课链路。");
  const pollingRef = useRef<number | null>(null);
  const stageRef = useRef<number | null>(null);

  const filteredCourses = useMemo(() => {
    const keyword = query.trim().toLowerCase();
    if (!keyword) return courses;
    return courses.filter((course) => {
      return course.Name.toLowerCase().includes(keyword) || String(course.ID).includes(keyword);
    });
  }, [courses, query]);

  const selectedRecord = useMemo(() => {
    if (!selectedCourse) return null;
    return selections.find((item) => item.course_id === selectedCourse.ID) || null;
  }, [selectedCourse, selections]);

  const clearTimers = useCallback(() => {
    if (pollingRef.current) window.clearInterval(pollingRef.current);
    if (stageRef.current) window.clearInterval(stageRef.current);
    pollingRef.current = null;
    stageRef.current = null;
  }, []);

  const loadSelections = useCallback(async () => {
    if (!auth) return;
    const payload = await apiRequest<Selection[]>("/auth/selections", auth.token);
    setSelections(payload.data || []);
  }, [auth]);

  const loadCourses = useCallback(async () => {
    setLoading(true);
    try {
      const payload = await apiRequest<Course[]>("/courses");
      const list = normalizeCourses(payload.data);
      setCourses(list);
      setSelectedCourse((current) => current || list[0] || null);
      setNotice(list.length > 0 ? "课程列表已加载，选择一门课程开始操作。" : "当前没有课程数据。");
    } catch (error) {
      setNotice(error instanceof Error ? error.message : "课程列表加载失败");
    } finally {
      setLoading(false);
    }
  }, []);

  const refreshCourse = useCallback(
    async (id: number) => {
      const payload = await apiRequest<Course>(`/courses/${id}`);
      if (payload.data) {
        setSelectedCourse(payload.data);
        setCourses((current) => current.map((course) => (course.ID === id ? payload.data! : course)));
      }
    },
    []
  );

  useEffect(() => {
    void loadCourses();
  }, [loadCourses]);

  useEffect(() => {
    if (auth) void loadSelections();
  }, [auth, loadSelections]);

  useEffect(() => {
    return () => clearTimers();
  }, [clearTimers]);

  async function handleAuth(event: FormEvent) {
    event.preventDefault();
    setLoading(true);
    try {
      if (authMode === "register") {
        await apiRequest("/register", undefined, {
          method: "POST",
          body: JSON.stringify({ sid, name, password })
        });
        setAuthMode("login");
        setNotice("注册成功，现在可以登录。");
        return;
      }

      const payload = await apiRequest("/login", undefined, {
        method: "POST",
        body: JSON.stringify({ sid, password })
      });
      if (!payload.token || !payload.name || !payload.id) {
        throw new Error("登录响应缺少 token");
      }
      const nextAuth = { token: payload.token, name: payload.name, id: payload.id };
      localStorage.setItem(tokenKey, payload.token);
      localStorage.setItem(userKey, JSON.stringify({ name: payload.name, id: payload.id }));
      setAuth(nextAuth);
      setNotice("登录成功，可以开始选课。");
    } catch (error) {
      setNotice(error instanceof Error ? error.message : "操作失败");
    } finally {
      setLoading(false);
    }
  }

  function logout() {
    clearTimers();
    localStorage.removeItem(tokenKey);
    localStorage.removeItem(userKey);
    setAuth(null);
    setSelections([]);
    setResultState("idle");
    setActiveStep(0);
    setNotice("已退出登录。");
  }

  function startStageAnimation() {
    setActiveStep(1);
    if (stageRef.current) window.clearInterval(stageRef.current);
    stageRef.current = window.setInterval(() => {
      setActiveStep((current) => Math.min(current + 1, timeline.length - 2));
    }, 520);
  }

  function stopPolling(finalState: ResultState) {
    clearTimers();
    setResultState(finalState);
    setActiveStep(finalState === "success" ? timeline.length : timeline.length - 1);
  }

  async function pollResult(courseID: number) {
    if (!auth) return;
    try {
      const payload = await apiRequest(`/auth/result/${courseID}`, auth.token);
      const nextState = resultFromMessage(payload.msg);
      if (nextState !== "pending") {
        stopPolling(nextState);
        setNotice(payload.msg || statusLabel(nextState));
        await Promise.all([refreshCourse(courseID), loadSelections()]);
      }
    } catch (error) {
      stopPolling("failed");
      setNotice(error instanceof Error ? error.message : "查询结果失败");
    }
  }

  async function selectCourse() {
    if (!auth || !selectedCourse) {
      setNotice("请先登录并选择课程。");
      return;
    }
    setBusyAction("select");
    setResultState("pending");
    startStageAnimation();
    try {
      const payload = await apiRequest(`/auth/select/${selectedCourse.ID}`, auth.token, { method: "POST" });
      setNotice(payload.msg || "排队中");
      await pollResult(selectedCourse.ID);
      pollingRef.current = window.setInterval(() => void pollResult(selectedCourse.ID), 1000);
    } catch (error) {
      stopPolling("failed");
      setNotice(error instanceof Error ? error.message : "选课失败");
    } finally {
      setBusyAction(null);
    }
  }

  async function dropCourse() {
    if (!auth || !selectedCourse) {
      setNotice("请先登录并选择课程。");
      return;
    }
    setBusyAction("drop");
    try {
      const payload = await apiRequest(`/auth/select/${selectedCourse.ID}`, auth.token, { method: "DELETE" });
      clearTimers();
      setResultState("dropped");
      setActiveStep(0);
      setNotice(payload.msg || "退课成功");
      await Promise.all([refreshCourse(selectedCourse.ID), loadSelections()]);
    } catch (error) {
      setNotice(error instanceof Error ? error.message : "退课失败");
    } finally {
      setBusyAction(null);
    }
  }

  return (
    <main className="shell">
      <header className="topbar">
        <div className="brand">
          <div className="brand-mark">
            <Sparkles size={22} />
          </div>
          <div>
            <h1>高并发选课演示台</h1>
            <p>Redis Lua + Stream + RabbitMQ + MySQL</p>
          </div>
        </div>
          <div className="top-actions">
          <div className="health-pill">
            <Activity size={16} />
            秒杀链路展示
          </div>
          {auth ? (
            <button className="user-pill" onClick={logout}>
              <UserRound size={16} />
              {auth.name}
              <LogOut size={16} />
            </button>
          ) : null}
        </div>
      </header>

      <section className="workspace">
        <aside className="course-panel">
          <div className="panel-heading">
            <div>
              <span className="eyebrow">Course Pool</span>
              <h2>课程池</h2>
            </div>
            <button className="icon-button" onClick={() => void loadCourses()} disabled={loading}>
              {loading ? <Loader2 className="spin" size={18} /> : <RotateCcw size={18} />}
            </button>
          </div>

          <label className="search-box">
            <Search size={17} />
            <input
              placeholder="搜索课程或 ID"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
            />
          </label>

          <div className="course-list">
            {filteredCourses.map((course) => (
              <button
                key={course.ID}
                className={`course-item ${selectedCourse?.ID === course.ID ? "active" : ""}`}
                onClick={() => {
                  setSelectedCourse(course);
                  setResultState("idle");
                  setActiveStep(0);
                  clearTimers();
                }}
              >
                <span className="course-main">
                  <span className="course-name">{course.Name}</span>
                  <span className="course-id">#{course.ID}</span>
                </span>
                <span className={`stock ${course.Stock > 0 ? "available" : "empty"}`}>库存 {course.Stock}</span>
              </button>
            ))}
          </div>
        </aside>

        <section className="center-panel">
          {!auth ? (
            <section className="auth-card">
              <div className="auth-copy">
                <span className="eyebrow">Student Access</span>
                <h2>{authMode === "login" ? "登录学生账号" : "注册学生账号"}</h2>
                <p>登录后可以发起选课、查询排队结果，并查看 Redis 到 MySQL 的异步链路演示。</p>
              </div>
              <form className="auth-form" onSubmit={handleAuth}>
                <input placeholder="学号" value={sid} onChange={(event) => setSid(event.target.value)} />
                {authMode === "register" ? (
                  <input placeholder="姓名" value={name} onChange={(event) => setName(event.target.value)} />
                ) : null}
                <input
                  type="password"
                  placeholder="密码"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                />
                <button className="primary-button" type="submit" disabled={loading}>
                  {loading ? <Loader2 className="spin" size={18} /> : null}
                  {authMode === "login" ? "登录" : "注册"}
                </button>
                <button
                  className="ghost-button"
                  type="button"
                  onClick={() => setAuthMode(authMode === "login" ? "register" : "login")}
                >
                  {authMode === "login" ? "没有账号，去注册" : "已有账号，去登录"}
                </button>
              </form>
            </section>
          ) : (
            <section className="operation-card">
              <div className="course-hero">
                <div>
                  <span className="eyebrow">Selected Course</span>
                  <h2>{selectedCourse?.Name || "请选择课程"}</h2>
                  <p>
                    教师 ID {selectedCourse?.TeacherID || "-"} · 课程 ID {selectedCourse?.ID || "-"}
                  </p>
                </div>
                <div className="stock-meter">
                  <span>实时库存</span>
                  <strong>{selectedCourse?.Stock ?? "-"}</strong>
                </div>
              </div>

              <div className="operation-grid">
                <div className="command-block">
                  <span className="eyebrow">Command</span>
                  <div className="action-row">
                    <button className="primary-button large" onClick={() => void selectCourse()} disabled={busyAction !== null}>
                      {busyAction === "select" ? <Loader2 className="spin" size={18} /> : <BookOpen size={18} />}
                      立即选课
                    </button>
                    <button className="danger-button large" onClick={() => void dropCourse()} disabled={busyAction !== null}>
                      {busyAction === "drop" ? <Loader2 className="spin" size={18} /> : <XCircle size={18} />}
                      退课
                    </button>
                  </div>
                </div>

                <div className={`result-banner ${resultState}`}>
                  <div>
                    <span>当前请求状态</span>
                    <strong>{statusLabel(resultState)}</strong>
                  </div>
                  <p>{notice}</p>
                </div>
              </div>

              <div className="selection-strip">
                <div className="strip-header">
                  <span className="eyebrow">My Selections</span>
                  <button className="ghost-button compact" onClick={() => void loadSelections()}>
                    刷新记录
                  </button>
                </div>
                <div className="selection-list">
                  {selections.length === 0 ? (
                    <p className="empty-text">暂无选课记录。</p>
                  ) : (
                    selections.slice(0, 5).map((item) => (
                      <div
                        key={item.selection_id}
                        className={`selection-row ${item.course_id === selectedRecord?.course_id ? "active" : ""}`}
                      >
                        <span>{item.course_name}</span>
                        <small>{item.status_text}</small>
                      </div>
                    ))
                  )}
                </div>
              </div>
            </section>
          )}

          <section className="flow-map">
            {timeline.map((step, index) => {
              const Icon = step.icon;
              const active = index < activeStep;
              const current = resultState === "pending" && index === activeStep - 1;
              return (
                <div className={`flow-node ${active ? "active" : ""} ${current ? "current" : ""}`} key={step.title}>
                  <div className="node-icon">
                    <Icon size={19} />
                  </div>
                  <div>
                    <strong>{step.title}</strong>
                    <span>{step.desc}</span>
                  </div>
                  {index < timeline.length - 1 ? <ArrowRight className="flow-arrow" size={18} /> : null}
                </div>
              );
            })}
          </section>
        </section>

        <aside className="timeline-panel">
          <div className="panel-heading">
            <div>
              <span className="eyebrow">Async Status</span>
              <h2>状态时间线</h2>
            </div>
            <Clock3 size={20} />
          </div>

          <div className="timeline-list">
            {timeline.map((step, index) => {
              const Icon = step.icon;
              const done = index < activeStep;
              const current = resultState === "pending" && index === activeStep - 1;
              return (
                <div className={`timeline-item ${done ? "done" : ""} ${current ? "current" : ""}`} key={step.title}>
                  <div className="timeline-dot">
                    <Icon size={16} />
                  </div>
                  <div>
                    <strong>{step.title}</strong>
                    <p>{step.desc}</p>
                  </div>
                </div>
              );
            })}
          </div>
        </aside>
      </section>
    </main>
  );
}

import { FormEvent, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { api, clearAuth, loadAuth, saveAuth } from "./api";
import { ConfirmDialog } from "./components/ConfirmDialog";
import { Navbar } from "./components/Navbar";
import { SelectionDrawer } from "./components/SelectionDrawer";
import { ArchitecturePage } from "./pages/ArchitecturePage";
import { CourseDetailPage } from "./pages/CourseDetailPage";
import { DashboardPage } from "./pages/DashboardPage";
import { LoginPage } from "./pages/LoginPage";
import { PerformancePage } from "./pages/PerformancePage";
import { SelectionsPage } from "./pages/SelectionsPage";
import type { AuthUser, Course, FilterKey, ProcessingState, RouteState, Selection, SelectionStatus } from "./types";
import { getCourseCapacity, getRoute, normalizeCourses, normalizeSelection, resultFromMessage, routeToPath } from "./utils";

export function App() {
  const [route, setRoute] = useState<RouteState>(() => getRoute());
  const [auth, setAuth] = useState<AuthUser | null>(() => loadAuth());
  const [sid, setSid] = useState("");
  const [password, setPassword] = useState("");
  const [notice, setNotice] = useState("");
  const [loadingAuth, setLoadingAuth] = useState(false);
  const [loadingCourses, setLoadingCourses] = useState(false);
  const [courses, setCourses] = useState<Course[]>([]);
  const [selections, setSelections] = useState<Selection[]>([]);
  const [query, setQuery] = useState("");
  const [filter, setFilter] = useState<FilterKey>("all");
  const [pendingCourseIds, setPendingCourseIds] = useState<Set<number>>(() => new Set());
  const [processingCourse, setProcessingCourse] = useState<Course | null>(null);
  const [processingState, setProcessingState] = useState<ProcessingState>("idle");
  const [activeStep, setActiveStep] = useState(0);
  const [processingMessage, setProcessingMessage] = useState("请求已进入异步队列");
  const [dropTarget, setDropTarget] = useState<number | null>(null);
  const stageRef = useRef<number | null>(null);
  const pollingRef = useRef<number | null>(null);
  const demoLoginRef = useRef(false);

  const navigate = useCallback((path: string) => {
    window.history.pushState({}, "", path);
    setRoute(getRoute(path));
  }, []);

  const clearTimers = useCallback(() => {
    if (stageRef.current) window.clearInterval(stageRef.current);
    if (pollingRef.current) window.clearInterval(pollingRef.current);
    stageRef.current = null;
    pollingRef.current = null;
  }, []);

  useEffect(() => {
    const handlePop = () => setRoute(getRoute());
    window.addEventListener("popstate", handlePop);
    return () => window.removeEventListener("popstate", handlePop);
  }, []);

  useEffect(() => {
    return () => clearTimers();
  }, [clearTimers]);

  const loadCourses = useCallback(async () => {
    setLoadingCourses(true);
    try {
      const payload = await api.getCourses();
      setCourses(normalizeCourses(payload.data));
      setNotice("");
    } catch (error) {
      setCourses(normalizeCourses([]));
      setNotice(error instanceof Error ? `课程接口暂不可用，已展示演示数据：${error.message}` : "课程接口暂不可用，已展示演示数据。");
    } finally {
      setLoadingCourses(false);
    }
  }, []);

  const loadSelections = useCallback(async () => {
    if (!auth) {
      setSelections([]);
      return;
    }
    try {
      const payload = await api.getSelections(auth.token);
      setSelections((payload.data || []).map(normalizeSelection));
    } catch (error) {
      setNotice(error instanceof Error ? error.message : "选课记录加载失败");
    }
  }, [auth]);

  useEffect(() => {
    void loadCourses();
  }, [loadCourses]);

  useEffect(() => {
    void loadSelections();
  }, [loadSelections]);

  const enterDemoMode = useCallback(async (redirectToPerformance = false) => {
    if (demoLoginRef.current) return;
    demoLoginRef.current = true;
    setLoadingAuth(true);
    setNotice("");
    try {
      const payload = await api.demoLogin();
      if (!payload.token || !payload.name || !payload.id) {
        throw new Error("演示登录响应缺少 token");
      }
      const nextAuth = { token: payload.token, name: payload.name, id: payload.id };
      saveAuth(nextAuth);
      setAuth(nextAuth);
      if (redirectToPerformance) navigate("/performance");
    } catch (error) {
      setNotice(error instanceof Error ? `演示登录失败：${error.message}` : "演示登录失败");
      if (route.page !== "login") navigate("/login");
    } finally {
      setLoadingAuth(false);
      demoLoginRef.current = false;
    }
  }, [navigate, route.page]);

  useEffect(() => {
    if (!auth && route.page !== "login") {
      void enterDemoMode(false);
    }
  }, [auth, enterDemoMode, route.page]);

  const selectedCourseIds = useMemo(() => new Set(selections.filter((item) => item.status === 1).map((item) => item.course_id)), [selections]);

  const getCourseStatus = useCallback(
    (course: Course): SelectionStatus | "available" | "hot" | "full" => {
      if (pendingCourseIds.has(course.ID)) return "pending";
      if (selectedCourseIds.has(course.ID)) return "success";
      if (course.Stock <= 0) return "full";
      const ratio = course.Stock / getCourseCapacity(course);
      if (ratio > 0 && ratio <= 0.2) return "hot";
      return "available";
    },
    [pendingCourseIds, selectedCourseIds]
  );

  const visibleCourses = useMemo(() => {
    const keyword = query.trim().toLowerCase();
    return courses.filter((course) => {
      const status = getCourseStatus(course);
      const matchesQuery =
        !keyword ||
        course.Name.toLowerCase().includes(keyword) ||
        String(course.ID).includes(keyword) ||
        String(course.TeacherID).includes(keyword);
      const matchesFilter =
        filter === "all" ||
        (filter === "available" && (status === "available" || status === "hot")) ||
        (filter === "hot" && status === "hot") ||
        (filter === "full" && status === "full") ||
        (filter === "selected" && status === "success");
      return matchesQuery && matchesFilter;
    });
  }, [courses, filter, getCourseStatus, query]);

  async function handleLogin(event: FormEvent) {
    event.preventDefault();
    setLoadingAuth(true);
    setNotice("");
    try {
      const payload = await api.login(sid, password);
      if (!payload.token || !payload.name || !payload.id) {
        throw new Error("登录响应缺少 token");
      }
      const nextAuth = { token: payload.token, name: payload.name, id: payload.id };
      saveAuth(nextAuth);
      setAuth(nextAuth);
      navigate("/performance");
    } catch (error) {
      setNotice(error instanceof Error ? error.message : "登录失败");
    } finally {
      setLoadingAuth(false);
    }
  }

  function handleDemoLogin() {
    void enterDemoMode(true);
  }

  function logout() {
    clearTimers();
    clearAuth();
    setAuth(null);
    setSelections([]);
    setPendingCourseIds(new Set());
    setProcessingCourse(null);
    setProcessingState("idle");
    setActiveStep(0);
    setProcessingMessage("请求已进入异步队列");
    navigate("/login");
  }

  function closeProcessingModal() {
    if (processingState === "pending") return;
    setProcessingState("idle");
    setProcessingCourse(null);
    setActiveStep(0);
    setProcessingMessage("请求已进入异步队列");
  }

  function beginProcessing(course: Course) {
    clearTimers();
    setProcessingCourse(course);
    setProcessingState("pending");
    setProcessingMessage("正在发起选课请求");
    setActiveStep(0);
    setPendingCourseIds((current) => new Set(current).add(course.ID));
  }

  async function finishProcessing(course: Course, state: ProcessingState, message: string) {
    clearTimers();
    setProcessingState(state);
    setProcessingMessage(message);
    setActiveStep(2);
    setPendingCourseIds((current) => {
      const next = new Set(current);
      next.delete(course.ID);
      return next;
    });
    await Promise.all([loadCourses(), loadSelections()]);
  }

  async function pollResult(course: Course) {
    if (!auth) return;
    try {
      const payload = await api.getResult(course.ID, auth.token);
      const nextState = resultFromMessage(payload.status || payload.msg);
      if (nextState === "success") {
        await finishProcessing(course, "success", "选课成功，MySQL 已确认落库。");
      } else if (nextState === "failed") {
        await finishProcessing(course, "failed", payload.msg || "库存不足 / 重复选课 / 队列处理失败");
      } else if (nextState === "dropped") {
        await finishProcessing(course, "success", "当前课程已退课。");
      } else {
        setActiveStep(1);
        setProcessingMessage("排队中，正在等待最终确认。");
      }
    } catch (error) {
      await finishProcessing(course, "failed", error instanceof Error ? error.message : "查询结果失败");
    }
  }

  async function selectCourse(course: Course) {
    if (!auth) {
      setNotice("请先登录再提交选课请求。");
      navigate("/login");
      return;
    }
    if (getCourseStatus(course) === "full") return;
    beginProcessing(course);
    try {
      const payload = await api.selectCourse(course.ID, auth.token);
      setProcessingMessage(payload.msg || "请求已进入异步队列");
      setActiveStep(1);
      await pollResult(course);
      pollingRef.current = window.setInterval(() => void pollResult(course), 1200);
    } catch (error) {
      await finishProcessing(course, "failed", error instanceof Error ? error.message : "库存不足，选课失败");
    }
  }

  async function confirmDropCourse() {
    if (!auth || dropTarget === null) return;
    const courseId = dropTarget;
    setDropTarget(null);
    try {
      await api.dropCourse(courseId, auth.token);
      setNotice("退课成功");
      await Promise.all([loadCourses(), loadSelections()]);
    } catch (error) {
      setNotice(error instanceof Error ? error.message : "退课失败");
    }
  }

  const courseForRoute = route.page === "course" ? courses.find((course) => course.ID === route.courseId) : undefined;

  return (
    <div className="app-shell">
      {route.page !== "login" ? <Navbar route={route} auth={auth} onNavigate={navigate} onLogout={logout} /> : null}
      {notice && route.page !== "login" ? <div className="global-notice">{notice}</div> : null}

      {route.page === "login" ? (
        <LoginPage
          sid={sid}
          password={password}
          loading={loadingAuth}
          notice={notice}
          onSid={setSid}
          onPassword={setPassword}
          onSubmit={handleLogin}
          onDemoLogin={handleDemoLogin}
        />
      ) : route.page === "selections" ? (
        <SelectionsPage selections={selections} onDrop={setDropTarget} />
      ) : route.page === "course" ? (
        <CourseDetailPage
          course={courseForRoute}
          status={courseForRoute ? getCourseStatus(courseForRoute) : "full"}
          processingState={processingState}
          activeStep={activeStep}
          onBack={() => navigate("/dashboard")}
          onSelect={selectCourse}
        />
      ) : route.page === "performance" ? (
        <PerformancePage />
      ) : route.page === "architecture" ? (
        <ArchitecturePage />
      ) : (
        <DashboardPage
          courses={visibleCourses}
          selections={selections}
          query={query}
          filter={filter}
          onQuery={setQuery}
          onFilter={setFilter}
          getStatus={getCourseStatus}
          onSelect={selectCourse}
          onDetail={(course) => navigate(routeToPath({ page: "course", courseId: course.ID }))}
        />
      )}

      {loadingCourses ? <div className="loading-ribbon">正在同步课程库存...</div> : null}
      <SelectionDrawer
        course={processingCourse}
        open={processingState !== "idle" && Boolean(processingCourse)}
        state={processingState}
        activeStep={activeStep}
        message={processingMessage}
        onClose={closeProcessingModal}
      />
      <ConfirmDialog
        open={dropTarget !== null}
        title="确认退课"
        desc="退课后库存会恢复，后续可以重新选择该课程。"
        onCancel={() => setDropTarget(null)}
        onConfirm={() => void confirmDropCourse()}
      />
    </div>
  );
}

const tokenKey = "course_select_token";

const messageEl = document.getElementById("message");
const userInfoEl = document.getElementById("user-info");
const courseBodyEl = document.getElementById("course-body");
const resultTextEl = document.getElementById("result-text");
const resultStatusEl = document.getElementById("result-status");
const logoutBtn = document.getElementById("logout-btn");
const courseCountEl = document.getElementById("course-count");
const availableCountEl = document.getElementById("available-count");
const pollingStateEl = document.getElementById("polling-state");

const activePollers = new Set();
const finalResults = new Map();

const fastPollIntervalMs = 800;
const slowPollIntervalMs = 1500;
const slowPollAfterMs = 6000;
const maxPollDurationMs = 15000;

function getToken() {
  return localStorage.getItem(tokenKey) || "";
}

function setMessage(text, isError = false) {
  messageEl.textContent = text;
  messageEl.className = `message ${isError ? "error" : "success"}`;
}

function setPollingStateText() {
  pollingStateEl.textContent = activePollers.size > 0 ? `查询中(${activePollers.size})` : "空闲";
}

function setResultStatus(status, text) {
  resultStatusEl.className = `status-badge ${status}`;
  resultStatusEl.textContent = text;
}

function normalizeResult(data) {
  const message = (data?.msg || "排队中").trim();
  if (message.includes("成功")) {
    return { status: "success", text: message, final: true };
  }
  if (message.includes("失败") || message.includes("不足")) {
    return { status: "failed", text: message, final: true };
  }
  return { status: "pending", text: "排队中", final: false };
}

function renderResult(courseID, result) {
  const previousFinal = finalResults.get(courseID);
  const next = previousFinal && !result.final ? previousFinal : result;

  if (next.final) {
    finalResults.set(courseID, next);
  }

  const statusText = next.status === "success" ? "成功" : next.status === "failed" ? "失败" : "排队中";
  setResultStatus(next.status, statusText);
  resultTextEl.textContent = `课程 ${courseID}：${next.text}`;
  return next;
}

async function request(path, options = {}) {
  const token = getToken();
  const headers = {
    "Content-Type": "application/json",
    ...(options.headers || {}),
  };

  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const res = await fetch(path, {
    ...options,
    headers,
  });

  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(data.msg || `请求失败(${res.status})`);
  }
  return data;
}

async function loadCourses() {
  try {
    const data = await request("/courses", { method: "GET" });
    const list = Array.isArray(data.data) ? data.data : [];

    courseCountEl.textContent = String(list.length);
    availableCountEl.textContent = String(list.filter((course) => Number(course.Stock) > 0).length);

    if (list.length === 0) {
      courseBodyEl.innerHTML = "<tr><td colspan='4'>暂无课程</td></tr>";
      return;
    }

    courseBodyEl.innerHTML = list
      .map(
        (course) => `
        <tr>
          <td>${course.ID}</td>
          <td>${course.Name || "未命名课程"}</td>
          <td>${course.Stock}</td>
          <td><button type="button" data-id="${course.ID}">抢课</button></td>
        </tr>`
      )
      .join("");
  } catch (err) {
    setMessage(err.message, true);
  }
}

async function selectCourse(courseID) {
  const data = await request(`/auth/select/${courseID}`, { method: "POST" });
  setMessage(data.msg || "已提交请求，请稍后查询结果");
}

async function queryResult(courseID) {
  const data = await request(`/auth/result/${courseID}`, { method: "GET" });
  return normalizeResult(data);
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function setSelectButtonLoading(button, isLoading) {
  if (!button) {
    return;
  }
  button.disabled = isLoading;
  button.classList.toggle("is-loading", isLoading);
  button.textContent = isLoading ? "排队中..." : "抢课";
}

async function startAutoPolling(courseID, button) {
  if (activePollers.has(courseID)) {
    setMessage(`课程 ${courseID} 正在查询结果，请稍候`);
    return;
  }

  activePollers.add(courseID);
  setPollingStateText();
  setSelectButtonLoading(button, true);
  setResultStatus("pending", "排队中");
  resultTextEl.textContent = `课程 ${courseID}：排队中`;

  const startedAt = Date.now();
  let interval = fastPollIntervalMs;

  try {
    while (Date.now() - startedAt <= maxPollDurationMs) {
      const current = renderResult(courseID, await queryResult(courseID));

      if (current.final) {
        setMessage(
          current.status === "success" ? `课程 ${courseID} 抢课成功` : `课程 ${courseID} 抢课失败`,
          current.status !== "success"
        );
        return;
      }

      if (Date.now() - startedAt > slowPollAfterMs) {
        interval = slowPollIntervalMs;
      }
      await sleep(interval);
    }

    setResultStatus("pending", "超时");
    resultTextEl.textContent = `课程 ${courseID}：排队中（查询超时，可稍后手动查询）`;
    setMessage("结果查询超时，请稍后手动查询", true);
  } catch (err) {
    setMessage(err.message, true);
  } finally {
    activePollers.delete(courseID);
    setPollingStateText();
    setSelectButtonLoading(button, false);
  }
}

document.getElementById("refresh-btn").addEventListener("click", loadCourses);

document.getElementById("result-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  const courseID = Number(document.getElementById("result-course-id").value);
  if (!courseID) {
    setMessage("请输入正确的课程ID", true);
    return;
  }

  try {
    const current = renderResult(courseID, await queryResult(courseID));
    if (current.status === "pending") {
      setMessage(`课程 ${courseID} 仍在排队`, false);
      return;
    }
    setMessage(`课程 ${courseID} 查询完成：${current.text}`, current.status !== "success");
  } catch (err) {
    setMessage(err.message, true);
  }
});

courseBodyEl.addEventListener("click", async (e) => {
  const target = e.target;
  if (!(target instanceof HTMLButtonElement)) {
    return;
  }

  const courseID = Number(target.dataset.id);
  if (!courseID) {
    setMessage("课程ID无效", true);
    return;
  }

  try {
    await selectCourse(courseID);
    await startAutoPolling(courseID, target);
  } catch (err) {
    setMessage(err.message, true);
  }
});

logoutBtn.addEventListener("click", () => {
  localStorage.removeItem(tokenKey);
  localStorage.removeItem("course_select_name");
  window.location.href = "/";
});

(function init() {
  if (!getToken()) {
    window.location.href = "/";
    return;
  }

  const name = localStorage.getItem("course_select_name") || "已登录用户";
  userInfoEl.textContent = `当前用户：${name}`;
  setPollingStateText();
  loadCourses();
})();


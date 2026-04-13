const tokenKey = "course_select_token";

const messageEl = document.getElementById("message");
const userInfoEl = document.getElementById("user-info");
const courseBodyEl = document.getElementById("course-body");
const resultTextEl = document.getElementById("result-text");
const logoutBtn = document.getElementById("logout-btn");
const activePollers = new Set();

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
    const list = data.data || [];

    if (!Array.isArray(list) || list.length === 0) {
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
  const data = await request(`/auth/select/${courseID}`, {
    method: "POST",
  });

  setMessage(data.msg || "已提交请求，请稍后查询结果");
}

async function queryResult(courseID) {
  const data = await request(`/auth/result/${courseID}`, {
    method: "GET",
  });

  const message = data.msg || "排队中";
  resultTextEl.textContent = `课程 ${courseID}：${message}`;
  return message;
}

function sleep(ms) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}

function setSelectButtonLoading(button, isLoading) {
  if (!button) {
    return;
  }

  button.disabled = isLoading;
  button.classList.toggle("is-loading", isLoading);
  button.textContent = isLoading ? "排队中..." : "抢课";
}

function isFinalResult(message) {
  return message === "抢课成功" || message === "抢课失败";
}

async function startAutoPolling(courseID, button) {
  if (activePollers.has(courseID)) {
    setMessage(`课程 ${courseID} 正在查询结果，请稍候`);
    return;
  }

  activePollers.add(courseID);
  setSelectButtonLoading(button, true);

  const startedAt = Date.now();
  let interval = fastPollIntervalMs;

  try {
    while (Date.now() - startedAt <= maxPollDurationMs) {
      const message = await queryResult(courseID);

      if (isFinalResult(message)) {
        setMessage(`课程 ${courseID}${message === "抢课成功" ? "处理完成" : "处理失败"}`, message !== "抢课成功");
        return;
      }

      if (Date.now() - startedAt > slowPollAfterMs) {
        interval = slowPollIntervalMs;
      }

      await sleep(interval);
    }

    resultTextEl.textContent = `课程 ${courseID}：排队中（查询超时，可稍后手动查询）`;
    setMessage("结果查询超时，请稍后手动查询", true);
  } catch (err) {
    setMessage(err.message, true);
  } finally {
    activePollers.delete(courseID);
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
    const message = await queryResult(courseID);
    setMessage(`课程 ${courseID}${message === "排队中" ? "仍在排队" : `查询完成：${message}`}`, false);
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
    resultTextEl.textContent = `课程 ${courseID}：排队中`;
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
  userInfoEl.textContent = name;
  loadCourses();
})();


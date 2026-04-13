const tokenKey = "course_select_token";

const messageEl = document.getElementById("message");
const userInfoEl = document.getElementById("user-info");
const courseBodyEl = document.getElementById("course-body");
const resultTextEl = document.getElementById("result-text");
const logoutBtn = document.getElementById("logout-btn");

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

  resultTextEl.textContent = data.msg || "排队中";
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
    await queryResult(courseID);
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


const tokenKey = "course_select_token";

const msgEl = document.getElementById("auth-message");
const loginForm = document.getElementById("login-form");
const registerForm = document.getElementById("register-form");
const loginTab = document.getElementById("tab-login");
const registerTab = document.getElementById("tab-register");

function setMessage(text, isError = false) {
  msgEl.textContent = text;
  msgEl.className = `message ${isError ? "error" : "success"}`;
}

function setActiveTab(isLogin) {
  loginTab.classList.toggle("active", isLogin);
  registerTab.classList.toggle("active", !isLogin);
  loginForm.classList.toggle("active", isLogin);
  registerForm.classList.toggle("active", !isLogin);
}

async function postJSON(path, payload) {
  const res = await fetch(path, {
	method: "POST",
	headers: { "Content-Type": "application/json" },
	body: JSON.stringify(payload),
  });

  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
	throw new Error(data.msg || `请求失败(${res.status})`);
  }
  return data;
}

loginTab.addEventListener("click", () => setActiveTab(true));
registerTab.addEventListener("click", () => setActiveTab(false));

loginForm.addEventListener("submit", async (e) => {
  e.preventDefault();
  const sid = document.getElementById("login-sid").value.trim();
  const password = document.getElementById("login-password").value;

  if (!sid || !password) {
	setMessage("请输入学号和密码", true);
	return;
  }

  try {
	const data = await postJSON("/login", { sid, password });
	if (!data.token) {
	  throw new Error("登录成功但未返回 token");
	}
	localStorage.setItem(tokenKey, data.token);
	localStorage.setItem("course_select_name", data.name || "用户");
	setMessage("登录成功，正在跳转...");
	window.location.href = "/home";
  } catch (err) {
	setMessage(err.message, true);
  }
});

registerForm.addEventListener("submit", async (e) => {
  e.preventDefault();
  const sid = document.getElementById("register-sid").value.trim();
  const name = document.getElementById("register-name").value.trim();
  const password = document.getElementById("register-password").value;

  if (!sid || !name || !password) {
	setMessage("请填写完整注册信息", true);
	return;
  }

  try {
	await postJSON("/register", { sid, name, password });
	setMessage("注册成功，请登录");
	registerForm.reset();
	setActiveTab(true);
  } catch (err) {
	setMessage(err.message, true);
  }
});

(function init() {
  if (localStorage.getItem(tokenKey)) {
	setMessage("检测到已有登录态，可直接进入主页");
  }
})();


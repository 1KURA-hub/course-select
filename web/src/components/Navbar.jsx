import { BookOpen, ClipboardList, LogOut } from "lucide-react";

export default function Navbar({ user, page, onNav, onLogout }) {
  return (
    <nav className="navbar">
      <div className="nav-brand">
        <span className="dot" />
        <span>高并发选课系统</span>
      </div>
      <div className="nav-links">
        <button className={page === "hall" ? "active" : ""} onClick={() => onNav("hall")}>
          <BookOpen size={16} /> 课程市场
        </button>
        <button className={page === "selections" ? "active" : ""} onClick={() => onNav("selections")}>
          <ClipboardList size={16} /> 我的选课
        </button>
      </div>
      <div className="nav-user">
        <span style={{ color: "var(--text-dim)" }}>{user?.name || user?.sid}</span>
        <button className="btn btn-ghost" style={{ padding: "6px 14px", fontSize: 12 }} onClick={onLogout}>
          <LogOut size={14} /> 退出
        </button>
      </div>
    </nav>
  );
}

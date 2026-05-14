import {
  BarChart3,
  BookOpen,
  GitBranch,
  GraduationCap,
  LayoutDashboard,
  LogOut,
  UserRound
} from "lucide-react";
import type { AuthUser, RouteState } from "../types";

const navItems: Array<{ page: RouteState["page"]; label: string; path: string; icon: typeof LayoutDashboard }> = [
  { page: "performance", label: "性能看板", path: "/performance", icon: BarChart3 },
  { page: "dashboard", label: "课程市场", path: "/dashboard", icon: LayoutDashboard },
  { page: "selections", label: "我的选课", path: "/selections", icon: BookOpen },
  { page: "architecture", label: "架构可视化", path: "/architecture", icon: GitBranch }
];

export function Navbar({
  route,
  auth,
  onNavigate,
  onLogout
}: {
  route: RouteState;
  auth: AuthUser | null;
  onNavigate: (path: string) => void;
  onLogout: () => void;
}) {
  return (
    <nav className="navbar">
      <button className="brand" onClick={() => onNavigate("/performance")}>
        <span className="brand-mark"><GraduationCap size={20} /></span>
        <span>
          <strong>CourseRush</strong>
          <small>高并发选课系统</small>
        </span>
      </button>

      <div className="nav-links">
        {navItems.map((item) => {
          const Icon = item.icon;
          const active = route.page === item.page || (route.page === "course" && item.page === "dashboard");
          return (
            <button key={item.path} className={active ? "active" : ""} onClick={() => onNavigate(item.path)}>
              <Icon size={16} />
              {item.label}
            </button>
          );
        })}
      </div>

      <div className="nav-user">
        {auth ? (
          <>
            <span className="user-chip"><UserRound size={14} /> {auth.name}</span>
            <button className="icon-button" onClick={onLogout} title="退出登录" aria-label="退出登录">
              <LogOut size={17} />
            </button>
          </>
        ) : (
          <button className="login-link" onClick={() => onNavigate("/login")}>登录</button>
        )}
      </div>
    </nav>
  );
}

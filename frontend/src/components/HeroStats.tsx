import { BookCheck, Flame, Lock, UserCheck } from "lucide-react";
import type { Course, Selection } from "../types";

export function HeroStats({ courses, selections }: { courses: Course[]; selections: Selection[] }) {
  const available = courses.filter((course) => course.Stock > 0).length;
  const hot = courses.filter((course) => course.Stock > 0 && course.Stock <= 5).length;
  const full = courses.filter((course) => course.Stock <= 0).length;
  const selected = selections.filter((item) => item.status === 1).length;
  const stats = [
    { label: "可选课程", value: available, icon: BookCheck, tone: "blue" },
    { label: "即将满员", value: hot, icon: Flame, tone: "orange" },
    { label: "已满课程", value: full, icon: Lock, tone: "muted" },
    { label: "我的选课数", value: selected, icon: UserCheck, tone: "green" }
  ];

  return (
    <div className="hero-stats">
      {stats.map((stat, index) => {
        const Icon = stat.icon;
        return (
          <div className={`stat-card ${stat.tone}`} key={stat.label} style={{ animationDelay: `${index * 80}ms` }}>
            <Icon size={18} />
            <strong>{stat.value}</strong>
            <span>{stat.label}</span>
          </div>
        );
      })}
    </div>
  );
}

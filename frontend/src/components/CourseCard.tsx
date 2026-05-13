import { ArrowUpRight, Loader2 } from "lucide-react";
import type { Course, SelectionStatus } from "../types";
import { getCourseCapacity } from "../utils";
import { LoadingBadge, StatusBadge } from "./StatusBadge";

export function CourseCard({
  course,
  index,
  status,
  onSelect,
  onDetail
}: {
  course: Course;
  index: number;
  status: SelectionStatus | "available" | "hot" | "full";
  onSelect: (course: Course) => void;
  onDetail: (course: Course) => void;
}) {
  const capacity = getCourseCapacity(course);
  const percent = Math.max(0, Math.min(100, (course.Stock / capacity) * 100));
  const disabled = status === "full" || status === "success" || status === "pending";
  const tone = status === "hot" ? "hot" : status === "full" ? "full" : status === "success" ? "selected" : "";
  const buttonText = status === "full" ? "已满" : status === "success" ? "已选" : status === "pending" ? "处理中" : "选课";

  return (
    <article className={`course-card ${tone}`} style={{ animationDelay: `${index * 55}ms` }}>
      <div className="course-card-head">
        <span className="course-code">COURSE {String(course.ID).padStart(3, "0")}</span>
        <button className="icon-button jump" onClick={() => onDetail(course)} title="查看课程详情" aria-label="查看课程详情">
          <ArrowUpRight size={16} />
        </button>
      </div>
      <h3>{course.Name}</h3>
      <p className="teacher">授课教师 ID: {course.TeacherID}</p>

      <div className="stock-line">
        <span>剩余名额</span>
        <strong>{course.Stock}</strong>
        <small>/ {capacity}</small>
      </div>
      <div className="stock-progress">
        <span style={{ width: `${percent}%` }} />
      </div>

      <div className="course-card-foot">
        {status === "pending" ? (
          <LoadingBadge />
        ) : status === "success" ? (
          <StatusBadge status="success" />
        ) : status === "full" ? (
          <span className="state-pill full">已满员</span>
        ) : status === "hot" ? (
          <span className="state-pill hot">即将满员</span>
        ) : (
          <span className="state-pill open">可选</span>
        )}
        <button className="primary-button compact course-action" disabled={disabled} onClick={() => onSelect(course)}>
          {status === "pending" ? <Loader2 className="spin" size={15} /> : null}
          {buttonText}
        </button>
      </div>
    </article>
  );
}

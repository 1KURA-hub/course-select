import { ArrowLeft, CheckCircle2 } from "lucide-react";
import type { CSSProperties } from "react";
import { ProcessingTimeline } from "../components/ProcessingTimeline";
import type { Course, ProcessingState, SelectionStatus } from "../types";
import { getCourseCapacity } from "../utils";

export function CourseDetailPage({
  course,
  status,
  processingState,
  activeStep,
  onBack,
  onSelect
}: {
  course: Course | undefined;
  status: SelectionStatus | "available" | "hot" | "full";
  processingState: ProcessingState;
  activeStep: number;
  onBack: () => void;
  onSelect: (course: Course) => void;
}) {
  if (!course) {
    return (
      <section className="page">
        <button className="ghost-button" onClick={onBack}><ArrowLeft size={16} /> 返回课程市场</button>
        <div className="empty-panel">未找到课程。</div>
      </section>
    );
  }
  const capacity = getCourseCapacity(course);
  const percent = Math.max(0, Math.min(100, (course.Stock / capacity) * 100));

  return (
    <section className="page">
      <button className="ghost-button back-button" onClick={onBack}><ArrowLeft size={16} /> 返回课程市场</button>
      <div className="course-detail-layout">
        <div className="detail-main-card">
          <span className="eyebrow">Course Detail</span>
          <h1>{course.Name}</h1>
          <p>教师 ID {course.TeacherID} · COURSE {String(course.ID).padStart(3, "0")}</p>
          <div className="detail-tags">
            <span>JWT 鉴权</span>
            <span>Redis Lua</span>
            <span>RabbitMQ</span>
            <span>MySQL</span>
          </div>
          <button className="primary-button" disabled={status === "full" || status === "success" || status === "pending"} onClick={() => onSelect(course)}>
            <CheckCircle2 size={16} />
            {status === "success" ? "已选" : status === "pending" ? "处理中" : status === "full" ? "已满" : "立即选课"}
          </button>
        </div>

        <div className="inventory-card">
          <div className="ring" style={{ "--percent": percent } as CSSProperties}>
            <strong>{course.Stock}</strong>
            <span>剩余名额</span>
          </div>
          <p>总容量 {capacity}</p>
        </div>
      </div>

      <div className="flow-card">
        <span className="eyebrow">High-Concurrency Flow</span>
        <h2>{"发起选课请求 -> 排队中 -> 结果确认"}</h2>
        <ProcessingTimeline activeStep={activeStep} state={processingState} />
      </div>
    </section>
  );
}

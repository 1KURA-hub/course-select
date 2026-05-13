import { Trash2 } from "lucide-react";
import { StatusBadge } from "../components/StatusBadge";
import type { Selection } from "../types";

export function SelectionsPage({
  selections,
  onDrop
}: {
  selections: Selection[];
  onDrop: (courseId: number) => void;
}) {
  const activeSelections = selections.filter((item) => item.status !== 2 && item.ui_status !== "dropped");

  return (
    <section className="page">
      <div className="page-heading">
        <span className="eyebrow">My Selections</span>
        <h1>我的选课</h1>
        <p>查看当前学生的排队、成功、失败和退课状态。</p>
      </div>

      <div className="selection-table">
        <div className="selection-row head">
          <span>课程名</span>
          <span>教师 ID</span>
          <span>状态</span>
          <span>选课时间</span>
          <span>操作</span>
        </div>
        {activeSelections.length === 0 ? (
          <div className="empty-panel">暂无选课记录。</div>
        ) : (
          activeSelections.map((item) => (
            <div className="selection-row" key={item.selection_id}>
              <strong>{item.course_name}</strong>
              <span>{item.teacher_id}</span>
              <StatusBadge status={item.ui_status || "pending"} label={item.status_text} />
              <span>{item.created_at || "当前会话"}</span>
              <button className="danger-button compact" disabled={item.status !== 1} onClick={() => onDrop(item.course_id)}>
                <Trash2 size={14} />
                退课
              </button>
            </div>
          ))
        )}
      </div>
    </section>
  );
}

import { useState, useEffect } from "react";
import { Trash2, Clock, CheckCircle2, XCircle, PackageOpen } from "lucide-react";
import { api } from "../api";
import { useToast } from "../App";

const statusMap = {
  1: { label: "已选课", cls: "status-selected", icon: CheckCircle2 },
  2: { label: "已退课", cls: "status-dropped", icon: XCircle },
};

export default function MySelections() {
  const [selections, setSelections] = useState([]);
  const [loading, setLoading] = useState(true);
  const addToast = useToast();

  useEffect(() => { load(); }, []);

  async function load() {
    try {
      const data = await api.getSelections();
      setSelections(data.selections || []);
    } catch (e) {
      addToast(e.message, "error");
    } finally {
      setLoading(false);
    }
  }

  async function handleDrop(courseId) {
    try {
      await api.dropCourse(courseId);
      addToast("退课成功", "success");
      load();
    } catch (e) {
      addToast(e.message, "error");
    }
  }

  return (
    <div className="selections-page">
      <h1>我的选课</h1>
      {loading ? (
        <div className="loading">
          {[...Array(3)].map((_, i) => <div key={i} className="skeleton" />)}
        </div>
      ) : selections.length === 0 ? (
        <div className="empty">
          <PackageOpen size={48} className="icon" />
          <p>还没有选课记录，去课程市场看看吧</p>
        </div>
      ) : (
        <div className="timeline">
          {selections.map((s) => {
            const st = statusMap[s.status] || { label: "未知", cls: "", icon: Clock };
            const Icon = st.icon;
            return (
              <div key={s.selection_id} className={`glass timeline-item ${s.status === 2 ? "dropped" : ""}`}>
                <span className={`status-badge ${st.cls}`}>
                  <Icon size={14} /> {st.label}
                </span>
                <div className="course-name">{s.course_name || `课程 #${s.course_id}`}</div>
                <div className="course-meta">
                  <span>课程 ID: {s.course_id}</span>
                  <span>教师 ID: {s.teacher_id}</span>
                </div>
                {s.status === 1 && (
                  <button className="btn btn-danger" style={{ marginTop: 14, padding: "8px 18px", fontSize: 13 }}
                    onClick={() => handleDrop(s.course_id)}>
                    <Trash2 size={14} /> 退课
                  </button>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

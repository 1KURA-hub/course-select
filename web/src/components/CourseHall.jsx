import { useState, useEffect } from "react";
import { Flame, AlertCircle } from "lucide-react";
import { api } from "../api";
import { useToast } from "../App";

export default function CourseHall({ onDetail, onSelecting }) {
  const [courses, setCourses] = useState([]);
  const [loading, setLoading] = useState(true);
  const addToast = useToast();

  useEffect(() => { loadCourses(); }, []);
  async function loadCourses() {
    try {
      const data = await api.getCourses();
      setCourses(data.courses || []);
    } catch (e) {
      addToast("加载课程失败: " + e.message, "error");
    } finally {
      setLoading(false);
    }
  }

  async function handleQuickSelect(e, course) {
    e.stopPropagation();
    if (course.stock <= 0) return;
    try {
      await api.selectCourse(course.id);
      onSelecting(course.id);
    } catch (err) {
      addToast(err.message, "error");
    }
  }

  const hotCount = courses.filter((c) => c.stock > 0 && c.stock <= 5).length;
  const fullCount = courses.filter((c) => c.stock <= 0).length;

  return (
    <div className="hall">
      <div className="hall-header">
        <h1>课程市场</h1>
        <div className="stats">
          <div className="stat-pill hot">
            <span className="dot" /> 即将满员 {hotCount}
          </div>
          <div className="stat-pill">
            <span className="dot" style={{ background: "var(--primary)", boxShadow: "0 0 6px var(--primary)" }} />
            可选 {courses.filter((c) => c.stock > 0).length}
          </div>
          <div className="stat-pill">
            <span className="dot" style={{ background: "var(--text-dim)", boxShadow: "none" }} />
            已满 {fullCount}
          </div>
        </div>
      </div>

      {loading ? (
        <div className="loading">
          {[...Array(6)].map((_, i) => <div key={i} className="skeleton" />)}
        </div>
      ) : (
        <div className="course-grid">
          {courses.map((c) => (
            <div key={c.id} className="glass course-card" onClick={() => onDetail(c.id)}>
              <div className="name">{c.name}</div>
              <div className="teacher">授课教师 ID: {c.teacher_id}</div>
              <div className="stock-row">
                <div className="stock-bar">
                  <div
                    className={`stock-fill ${c.stock <= 0 ? "gone" : c.stock <= 5 ? "low" : "full"}`}
                    style={{ width: `${Math.min((c.stock / 50) * 100, 100)}%` }}
                  />
                </div>
                <div className="stock-num" style={{ color: c.stock <= 0 ? "var(--danger)" : c.stock <= 5 ? "var(--warning)" : "var(--primary)" }}>
                  <span>{c.stock}</span> 剩余
                </div>
              </div>
              {c.stock > 0 && c.stock <= 5 && <span className="tag tag-almost"><AlertCircle size={12} /> 即将满员</span>}
              {c.stock <= 0 && <span className="tag tag-full">已售罄</span>}
              {c.stock > 5 && <span className="tag tag-full"><Flame size={12} /> 可选</span>}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

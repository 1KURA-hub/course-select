import { useState, useEffect } from "react";
import { X, ShoppingCart, Trash2 } from "lucide-react";
import { api } from "../api";
import { useToast } from "../App";

export default function CourseDetail({ id, onClose, onSelecting }) {
  const [course, setCourse] = useState(null);
  const [loading, setLoading] = useState(true);
  const addToast = useToast();

  useEffect(() => {
    api.getCourse(id).then((d) => setCourse(d.data)).catch((e) => addToast(e.message, "error")).finally(() => setLoading(false));
  }, [id]);

  async function handleSelect() {
    try {
      await api.selectCourse(id);
      addToast("已提交选课请求，排队中...", "info");
      onClose();
      onSelecting(id);
    } catch (e) {
      addToast(e.message, "error");
    }
  }

  async function handleDrop() {
    try {
      await api.dropCourse(id);
      addToast("退课成功", "success");
      onClose();
    } catch (e) {
      addToast(e.message, "error");
    }
  }

  function handleOverlay(e) { if (e.target === e.currentTarget) onClose(); }

  const stock = course?.Stock ?? 0;
  const total = 50;
  const pct = Math.min((stock / total) * 100, 100);
  const circumference = 2 * Math.PI * 34;
  const offset = circumference - (pct / 100) * circumference;
  const strokeColor = stock <= 0 ? "var(--danger)" : stock <= 5 ? "var(--warning)" : "var(--primary)";

  return (
    <div className="detail-overlay" onClick={handleOverlay}>
      <div className="glass detail-card">
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginBottom: 16 }}>
          {loading ? <div className="skeleton" style={{ width: "100%", height: 48 }} /> : (
            <>
              <div>
                <h2>{course?.Name}</h2>
                <p className="teacher">授课教师 ID: {course?.TeacherID}</p>
              </div>
              <button className="btn btn-ghost" style={{ padding: 8, minWidth: 36 }} onClick={onClose}>
                <X size={18} />
              </button>
            </>
          )}
        </div>

        {!loading && course && (
          <>
            <div className="stock-meter">
              <div className="stock-ring">
                <svg width="80" height="80" viewBox="0 0 80 80">
                  <circle className="bg" cx="40" cy="40" r="34" />
                  <circle className="fill" cx="40" cy="40" r="34"
                    stroke={strokeColor}
                    strokeDasharray={circumference}
                    strokeDashoffset={offset}
                  />
                </svg>
                <div className="center" style={{ color: strokeColor }}>{stock}</div>
              </div>
              <div className="stock-info">
                <div className="label">实时库存</div>
                <div className={`value ${stock <= 0 ? "danger" : stock <= 5 ? "warn" : ""}`}>
                  {stock <= 0 ? "已售罄" : stock <= 5 ? "库存紧张" : "库存充足"}
                </div>
              </div>
            </div>
            <div className="detail-actions">
              <button className="btn btn-primary" disabled={stock <= 0} onClick={handleSelect}>
                <ShoppingCart size={18} /> {stock <= 0 ? "已售罄" : "立即选课"}
              </button>
              <button className="btn btn-danger" onClick={handleDrop}>
                <Trash2 size={18} /> 退课
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}

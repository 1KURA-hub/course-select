import { CheckCircle, XCircle, AlertTriangle, Info } from "lucide-react";

const icons = { success: CheckCircle, error: XCircle, warning: AlertTriangle, info: Info };

export default function Toast({ toasts }) {
  if (!toasts.length) return null;
  return (
    <div className="toast-container">
      {toasts.map((t) => {
        const Icon = icons[t.type] || Info;
        return (
          <div key={t.id} className={`toast toast-${t.type}`}>
            <Icon size={18} />
            <span>{t.msg}</span>
          </div>
        );
      })}
    </div>
  );
}

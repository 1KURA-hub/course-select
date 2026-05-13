import { CheckCircle2, Clock3, Loader2, XCircle } from "lucide-react";
import type { SelectionStatus } from "../types";

const labels: Record<SelectionStatus, string> = {
  pending: "排队中",
  success: "已选",
  failed: "选课失败",
  dropped: "已退课"
};

export function StatusBadge({ status, label }: { status: SelectionStatus; label?: string }) {
  const Icon = status === "success" ? CheckCircle2 : status === "failed" ? XCircle : status === "dropped" ? XCircle : Clock3;
  return (
    <span className={`status-badge ${status}`}>
      {status === "pending" ? <span className="breathing-dot" /> : <Icon size={14} />}
      {label || labels[status]}
    </span>
  );
}

export function LoadingBadge({ text = "处理中" }: { text?: string }) {
  return (
    <span className="status-badge pending">
      <Loader2 className="spin" size={14} />
      {text}
    </span>
  );
}

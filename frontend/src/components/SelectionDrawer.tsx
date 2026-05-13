import { X } from "lucide-react";
import type { Course, ProcessingState } from "../types";
import { ProcessingTimeline } from "./ProcessingTimeline";

export function SelectionDrawer({
  course,
  open,
  state,
  activeStep,
  message,
  onClose
}: {
  course: Course | null;
  open: boolean;
  state: ProcessingState;
  activeStep: number;
  message: string;
  onClose: () => void;
}) {
  if (!open || !course) return null;
  const success = state === "success";
  const failed = state === "failed";

  return (
    <div className="drawer-backdrop" onClick={(event) => event.target === event.currentTarget && onClose()}>
      <section className="selection-modal">
        <div className="drawer-head">
          <div>
            <span className="eyebrow">Async Course Selection</span>
            <h2>{course.Name}</h2>
            <p>请求已进入高并发异步处理链路</p>
          </div>
          <button className="icon-button modal-close" onClick={onClose} aria-label="关闭处理面板">
            <X size={18} />
          </button>
        </div>

        <div className="modal-progress">
          <span>{Math.min(100, Math.round(((activeStep + (state === "success" ? 1 : 0)) / 6) * 100))}%</span>
          <div><i style={{ width: `${Math.min(100, Math.round(((activeStep + (state === "success" ? 1 : 0)) / 6) * 100))}%` }} /></div>
          <strong>{success ? "选课成功" : failed ? "选课失败" : message}</strong>
        </div>

        <ProcessingTimeline activeStep={activeStep} state={state} />

        {state !== "failed" ? (
          <div className={`drawer-result ${state}`}>
            <strong>{success ? "选课成功" : "请求已进入异步队列"}</strong>
            <p>{message}</p>
          </div>
        ) : null}
      </section>
    </div>
  );
}

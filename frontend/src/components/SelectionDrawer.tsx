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
  const progress = Math.min(100, Math.round(((activeStep + 1) / 3) * 100));

  return (
    <div className="drawer-backdrop" onClick={(event) => event.target === event.currentTarget && onClose()}>
      <section className="selection-modal">
        <div className="drawer-head">
          <div>
            <span className="eyebrow">Async Course Selection</span>
            <h2>{course.Name}</h2>
            <p>跟踪本次选课请求的业务处理状态</p>
          </div>
          <button className="icon-button modal-close" onClick={onClose} aria-label="关闭处理面板">
            <X size={18} />
          </button>
        </div>

        <div className="modal-progress">
          <span>{progress}%</span>
          <div><i style={{ width: `${progress}%` }} /></div>
          <strong>{success ? "选课成功" : failed ? "选课失败" : message}</strong>
        </div>

        <ProcessingTimeline activeStep={activeStep} state={state} message={message} />

        {state !== "pending" ? (
          <div className={`drawer-result ${state}`}>
            <strong>{success ? "选课成功" : "选课失败"}</strong>
            <p>{message}</p>
          </div>
        ) : null}
      </section>
    </div>
  );
}

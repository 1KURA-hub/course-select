import { CheckCircle2, Loader2, XCircle } from "lucide-react";
import { processingSteps } from "../data";
import type { ProcessingState } from "../types";

export function ProcessingTimeline({
  activeStep,
  state
}: {
  activeStep: number;
  state: ProcessingState;
}) {
  return (
    <div className="processing-timeline">
      {processingSteps.map((step, index) => {
        const Icon = step.icon;
        const done = index < activeStep || state === "success";
        const active = state === "pending" && index === activeStep;
        const failed = state === "failed" && index === activeStep;
        const blocked = state === "failed" && index > activeStep;
        return (
          <div className={`process-node ${done ? "done" : ""} ${active ? "active" : ""} ${failed ? "failed" : ""} ${blocked ? "blocked" : ""}`} key={step.title}>
            <div className="process-icon">
              {active ? <Loader2 className="spin" size={18} /> : failed ? <XCircle size={18} /> : done ? <CheckCircle2 size={18} /> : <Icon size={18} />}
            </div>
            <div>
              <strong>{step.title}</strong>
              <p>{failed ? "库存不足 / 重复选课 / 队列处理失败" : step.desc}</p>
            </div>
          </div>
        );
      })}
    </div>
  );
}

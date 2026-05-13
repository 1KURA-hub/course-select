import { AlertTriangle } from "lucide-react";

export function ConfirmDialog({
  title,
  desc,
  open,
  onCancel,
  onConfirm
}: {
  title: string;
  desc: string;
  open: boolean;
  onCancel: () => void;
  onConfirm: () => void;
}) {
  if (!open) return null;
  return (
    <div className="modal-backdrop">
      <div className="confirm-dialog">
        <AlertTriangle size={28} />
        <h3>{title}</h3>
        <p>{desc}</p>
        <div className="dialog-actions">
          <button className="ghost-button" onClick={onCancel}>取消</button>
          <button className="danger-button" onClick={onConfirm}>确认退课</button>
        </div>
      </div>
    </div>
  );
}

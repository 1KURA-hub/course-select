import { useEffect, useMemo, useState } from "react";

function parseMetric(value: string) {
  if (!/\d/.test(value)) return null;
  const numeric = Number(value.replace(/[^0-9.]/g, ""));
  if (!Number.isFinite(numeric)) return null;
  return {
    numeric,
    prefix: value.match(/^[^0-9]*/)?.[0] || "",
    suffix: value.match(/[^0-9.]*$/)?.[0] || ""
  };
}

export function MetricCard({ label, value, tone }: { label: string; value: string; tone: string }) {
  const parsed = useMemo(() => parseMetric(value), [value]);
  const [display, setDisplay] = useState(value);

  useEffect(() => {
    if (!parsed) {
      setDisplay(value);
      return;
    }

    let frame = 0;
    const frames = 30;
    const timer = window.setInterval(() => {
      frame += 1;
      const next = parsed.numeric * Math.min(frame / frames, 1);
      const formatted = parsed.numeric >= 1000 ? Math.round(next).toLocaleString() : Math.round(next).toString();
      setDisplay(`${parsed.prefix}${formatted}${parsed.suffix}`);
      if (frame >= frames) window.clearInterval(timer);
    }, 16);

    return () => window.clearInterval(timer);
  }, [parsed, value]);

  return (
    <div className={`metric-card ${tone}`}>
      <span>{label}</span>
      <strong>{display}</strong>
    </div>
  );
}

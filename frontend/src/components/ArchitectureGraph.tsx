import { useState } from "react";
import { architectureNodes } from "../data";

const layers = [
  { title: "接入层", tone: "indigo", nodes: ["Browser", "React Frontend", "Gin API", "JWT Auth"] },
  { title: "缓存层", tone: "cyan", nodes: ["Bloom Filter", "Redis Lua", "Redis Stream"] },
  { title: "队列层", tone: "amber", nodes: ["RabbitMQ", "Retry Queue", "Dead Letter Queue"] },
  { title: "持久层", tone: "green", nodes: ["MySQL", "Result Polling"] }
];

export function ArchitectureGraph() {
  const [hoveredLayer, setHoveredLayer] = useState<string | null>(null);

  return (
    <div className="architecture-graph">
      {layers.map((layer, layerIndex) => (
        <div
          className={`arch-layer ${layer.tone}`}
          key={layer.title}
          onMouseEnter={() => setHoveredLayer(layer.title)}
          onMouseLeave={() => setHoveredLayer(null)}
        >
          <div className="arch-layer-title">{layer.title}</div>
          <div className="arch-layer-nodes">
            {layer.nodes.map((label) => {
              const node = architectureNodes.find((item) => item.label === label)!;
              const Icon = node.icon;
              const active = hoveredLayer === layer.title;
              const dim = Boolean(hoveredLayer) && !active;
              return (
                <div
                  className={`arch-node ${dim ? "dim" : ""} ${active ? "layer-active" : ""}`}
                  title={node.desc}
                  key={node.label}
                >
                  <Icon size={22} />
                  <strong>{node.label}</strong>
                  <span>{node.desc}</span>
                </div>
              );
            })}
          </div>
          {layerIndex < layers.length - 1 ? (
            <svg className="layer-arrow" viewBox="0 0 100 34" preserveAspectRatio="none" aria-hidden="true">
              <path d="M50 1 C50 12 50 20 50 31" />
              <path d="M44 24 L50 32 L56 24" />
            </svg>
          ) : null}
        </div>
      ))}
    </div>
  );
}

import { ArchitectureGraph } from "../components/ArchitectureGraph";

export function ArchitecturePage() {
  return (
    <section className="page architecture-page">
      <div className="page-heading">
        <span className="eyebrow">System Architecture</span>
        <h1>架构可视化</h1>
        <p>从浏览器请求到 Redis 预扣、RabbitMQ 削峰和 MySQL 落库的完整链路。</p>
      </div>
      <ArchitectureGraph />
    </section>
  );
}

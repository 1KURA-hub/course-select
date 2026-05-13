import {
  Activity,
  Blocks,
  CheckCircle2,
  Database,
  Filter,
  KeyRound,
  MessageSquare,
  Monitor,
  Rabbit,
  RotateCcw,
  Server,
  ShieldCheck
} from "lucide-react";
import type { Course } from "./types";

export const courseNameFallbacks = [
  "高并发系统设计",
  "分布式缓存实战",
  "Go 后端工程化",
  "Redis 与消息队列",
  "MySQL 事务与索引优化",
  "微服务架构设计",
  "云原生应用开发",
  "操作系统原理",
  "数据结构与算法",
  "计算机网络"
];

export const fallbackCourses: Course[] = courseNameFallbacks.map((Name, index) => ({
  ID: index + 1,
  Name,
  Stock: [38, 8, 26, 5, 0, 17, 44, 12, 2, 0][index],
  TeacherID: 1001 + index
}));

export const processingSteps = [
  { title: "JWT 鉴权", desc: "校验学生身份与请求合法性", icon: KeyRound },
  { title: "Redis Lua 原子预扣库存", desc: "库存扣减与重复请求判断在 Redis 内完成", icon: ShieldCheck },
  { title: "Redis Stream 写入 Outbox", desc: "记录预扣成功消息，避免请求丢失", icon: MessageSquare },
  { title: "RabbitMQ 异步削峰", desc: "消费端平滑处理高峰请求", icon: Rabbit },
  { title: "MySQL 最终落库", desc: "事务扣减真实库存并写入选课记录", icon: Database },
  { title: "结果确认", desc: "前端轮询确认最终选课结果", icon: CheckCircle2 }
];

export const performanceMetrics = [
  { label: "QPS", value: "6,420", tone: "blue" },
  { label: "平均延迟", value: "18ms", tone: "green" },
  { label: "P99 延迟", value: "96ms", tone: "orange" },
  { label: "成功请求", value: "200,000", tone: "green" },
  { label: "失败请求", value: "1,284", tone: "red" },
  { label: "不超卖验证", value: "通过", tone: "blue" }
];

export const architectureNodes = [
  { label: "Browser", desc: "学生发起抢课请求", icon: Monitor },
  { label: "React Frontend", desc: "展示库存、状态与异步链路", icon: Blocks },
  { label: "Gin API", desc: "统一入口与路由处理", icon: Server },
  { label: "JWT Auth", desc: "无状态用户鉴权", icon: KeyRound },
  { label: "Bloom Filter", desc: "拦截不存在课程 ID", icon: Filter },
  { label: "Redis Lua", desc: "原子预扣库存与防重复", icon: ShieldCheck },
  { label: "Redis Stream", desc: "Outbox 消息缓冲", icon: MessageSquare },
  { label: "RabbitMQ", desc: "异步削峰与消费确认", icon: Rabbit },
  { label: "Retry Queue", desc: "失败消息延迟重试", icon: RotateCcw },
  { label: "Dead Letter Queue", desc: "异常消息兜底补偿", icon: Activity },
  { label: "MySQL", desc: "最终一致落库", icon: Database },
  { label: "Result Polling", desc: "前端轮询选课结果", icon: CheckCircle2 }
];

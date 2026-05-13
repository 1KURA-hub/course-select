import { Zap } from "lucide-react";
import { CourseGrid } from "../components/CourseGrid";
import { HeroStats } from "../components/HeroStats";
import { SearchFilterBar } from "../components/SearchFilterBar";
import type { Course, FilterKey, Selection, SelectionStatus } from "../types";

export function DashboardPage({
  courses,
  selections,
  query,
  filter,
  onQuery,
  onFilter,
  getStatus,
  onSelect,
  onDetail
}: {
  courses: Course[];
  selections: Selection[];
  query: string;
  filter: FilterKey;
  onQuery: (value: string) => void;
  onFilter: (value: FilterKey) => void;
  getStatus: (course: Course) => SelectionStatus | "available" | "hot" | "full";
  onSelect: (course: Course) => void;
  onDetail: (course: Course) => void;
}) {
  return (
    <section className="page dashboard-page">
      <div className="hero-panel">
        <div>
          <span className="eyebrow"><Zap size={14} /> 实时选课控制台</span>
          <h1>课程市场</h1>
          <p>查看课程余量，提交选课请求，并跟踪异步处理链路。</p>
        </div>
        <HeroStats courses={courses} selections={selections} />
      </div>

      <SearchFilterBar query={query} filter={filter} onQuery={onQuery} onFilter={onFilter} />
      <CourseGrid courses={courses} getStatus={getStatus} onSelect={onSelect} onDetail={onDetail} />
    </section>
  );
}

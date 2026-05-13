import { Search } from "lucide-react";
import type { FilterKey } from "../types";

const filters: Array<{ key: FilterKey; label: string }> = [
  { key: "all", label: "全部" },
  { key: "available", label: "可选" },
  { key: "hot", label: "即将满员" },
  { key: "full", label: "已满员" },
  { key: "selected", label: "我的已选" }
];

export function SearchFilterBar({
  query,
  filter,
  onQuery,
  onFilter
}: {
  query: string;
  filter: FilterKey;
  onQuery: (value: string) => void;
  onFilter: (value: FilterKey) => void;
}) {
  return (
    <div className="search-filter">
      <label className="search-box">
        <Search size={18} />
        <input
          value={query}
          onChange={(event) => onQuery(event.target.value)}
          placeholder="按课程名称、课程 ID、教师 ID 搜索"
        />
      </label>
      <div className="filter-tabs">
        {filters.map((item) => (
          <button key={item.key} className={filter === item.key ? "active" : ""} onClick={() => onFilter(item.key)}>
            {item.label}
          </button>
        ))}
      </div>
    </div>
  );
}

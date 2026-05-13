import type { Course, SelectionStatus } from "../types";
import { CourseCard } from "./CourseCard";

export function CourseGrid({
  courses,
  getStatus,
  onSelect,
  onDetail
}: {
  courses: Course[];
  getStatus: (course: Course) => SelectionStatus | "available" | "hot" | "full";
  onSelect: (course: Course) => void;
  onDetail: (course: Course) => void;
}) {
  if (courses.length === 0) {
    return <div className="empty-panel">没有匹配的课程。</div>;
  }

  return (
    <div className="course-grid">
      {courses.map((course, index) => (
        <CourseCard
          key={course.ID}
          course={course}
          index={index}
          status={getStatus(course)}
          onSelect={onSelect}
          onDetail={onDetail}
        />
      ))}
    </div>
  );
}

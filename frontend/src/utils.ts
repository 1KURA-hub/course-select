import { courseNameFallbacks, fallbackCourses } from "./data";
import type { Course, RouteState, Selection, SelectionStatus } from "./types";

const mojibakePattern = /Ã|Â|é|ç|è|å|æ|ï¿½|�/;
const defaultCapacity = 50;

export function isMojibake(value: string) {
  return mojibakePattern.test(value);
}

export function fallbackCourseName(id: number) {
  return courseNameFallbacks[(Math.max(id, 1) - 1) % courseNameFallbacks.length];
}

export function normalizeCourse(input: Partial<Course> & Record<string, unknown>, index = 0): Course {
  const id = Number(input.ID ?? input.id ?? index + 1);
  const rawName = String(input.Name ?? input.name ?? "");
  const name = !rawName || isMojibake(rawName) ? fallbackCourseName(id) : rawName;
  return {
    ID: id,
    Name: name,
    Stock: Number(input.Stock ?? input.stock ?? 0),
    TeacherID: Number(input.TeacherID ?? input.teacher_id ?? input.teacherId ?? 1000 + id)
  };
}

export function normalizeCourses(data: unknown): Course[] {
  if (!Array.isArray(data) || data.length === 0) return fallbackCourses;
  return data.map((item, index) => normalizeCourse(item as Partial<Course> & Record<string, unknown>, index));
}

export function normalizeSelection(item: Selection): Selection {
  const status = selectionStatusFromCode(item.status);
  return {
    ...item,
    course_name: !item.course_name || isMojibake(item.course_name)
      ? fallbackCourseName(item.course_id)
      : item.course_name,
    ui_status: status,
    created_at: item.created_at || "当前会话"
  };
}

export function selectionStatusFromCode(status: number): SelectionStatus {
  if (status === 1) return "success";
  if (status === 2) return "dropped";
  return "pending";
}

export function resultFromMessage(message?: string): SelectionStatus {
  if (!message) return "pending";
  if (message.includes("成功") || message === "success") return "success";
  if (message.includes("失败") || message === "failed") return "failed";
  if (message.includes("退课") || message === "dropped") return "dropped";
  return "pending";
}

export function getCourseCapacity(course: Course) {
  return Math.max(defaultCapacity, course.Stock);
}

export function getRoute(pathname = window.location.pathname): RouteState {
  if (pathname === "/login") return { page: "login" };
  if (pathname === "/") return { page: "performance" };
  if (pathname === "/dashboard") return { page: "dashboard" };
  if (pathname === "/selections") return { page: "selections" };
  if (pathname === "/performance") return { page: "performance" };
  if (pathname === "/architecture") return { page: "architecture" };
  const match = pathname.match(/^\/courses\/(\d+)$/);
  if (match) return { page: "course", courseId: Number(match[1]) };
  return { page: "performance" };
}

export function routeToPath(route: RouteState) {
  if (route.page === "login") return "/login";
  if (route.page === "dashboard") return "/dashboard";
  if (route.page === "selections") return "/selections";
  if (route.page === "performance") return "/performance";
  if (route.page === "architecture") return "/architecture";
  return `/courses/${route.courseId}`;
}

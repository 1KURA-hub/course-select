export type Course = {
  ID: number;
  Name: string;
  Stock: number;
  TeacherID: number;
};

export type SelectionStatus = "pending" | "success" | "failed" | "dropped";

export type Selection = {
  selection_id: number;
  student_id: number;
  course_id: number;
  status: number;
  status_text: string;
  course_name: string;
  teacher_id: number;
  created_at?: string;
  ui_status?: SelectionStatus;
};

export type ApiResponse<T = unknown> = {
  code?: number;
  msg?: string;
  data?: T;
  token?: string;
  name?: string;
  id?: number;
  course_id?: number;
  status?: string;
};

export type AuthUser = {
  token: string;
  name: string;
  id: number;
};

export type FilterKey = "all" | "available" | "hot" | "full" | "selected";

export type RouteState =
  | { page: "login" }
  | { page: "dashboard" }
  | { page: "selections" }
  | { page: "course"; courseId: number }
  | { page: "performance" }
  | { page: "architecture" };

export type ProcessingState = "idle" | "pending" | "success" | "failed";

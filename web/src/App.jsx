import { useState, useCallback, createContext, useContext } from "react";
import LoginPage from "./components/LoginPage";
import CourseHall from "./components/CourseHall";
import CourseDetail from "./components/CourseDetail";
import MySelections from "./components/MySelections";
import SelectionStatus from "./components/SelectionStatus";
import Toast from "./components/Toast";
import Navbar from "./components/Navbar";

export const ToastContext = createContext(null);
export const useToast = () => useContext(ToastContext);

export default function App() {
  const [user, setUser] = useState(() => {
    const saved = localStorage.getItem("user");
    return saved ? JSON.parse(saved) : null;
  });
  const [page, setPage] = useState("hall");
  const [detailId, setDetailId] = useState(null);
  const [selectingId, setSelectingId] = useState(null);
  const [toasts, setToasts] = useState([]);

  const addToast = useCallback((msg, type = "info") => {
    const id = Date.now();
    setToasts((prev) => [...prev, { id, msg, type }]);
    setTimeout(() => setToasts((prev) => prev.filter((t) => t.id !== id)), 3500);
  }, []);

  const handleLogin = (u) => {
    localStorage.setItem("user", JSON.stringify(u));
    setUser(u);
  };

  const handleLogout = () => {
    localStorage.removeItem("user");
    localStorage.removeItem("token");
    setUser(null);
  };

  if (!user) return <LoginPage onLogin={handleLogin} />;

  return (
    <ToastContext.Provider value={addToast}>
      <Navbar user={user} page={page} onNav={setPage} onLogout={handleLogout} />
      {page === "hall" && (
        <CourseHall onDetail={setDetailId} onSelecting={setSelectingId} />
      )}
      {page === "selections" && <MySelections />}
      {detailId && (
        <CourseDetail
          id={detailId}
          onClose={() => setDetailId(null)}
          onSelecting={setSelectingId}
        />
      )}
      {selectingId && (
        <SelectionStatus
          courseId={selectingId}
          onDone={() => setSelectingId(null)}
        />
      )}
      <Toast toasts={toasts} />
    </ToastContext.Provider>
  );
}

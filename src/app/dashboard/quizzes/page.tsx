"use client";

import { useEffect, useState } from "react";
import { useAuth } from "@/context/auth-context";
import { apiFetch } from "@/lib/api";
import MyQuizzesView from "../_components/MyQuizzesView";
import { type LibraryQuiz } from "../_components/QuizLibrary";

type QuizDTO = {
  id: string;
  title: string;
  category: string;
  difficulty: string;
  tags: string[];
  coverImageUrl: string | null;
  archived: boolean;
};

export default function MyQuizzesPage() {
  const { user } = useAuth();
  const [quizzes, setQuizzes] = useState<LibraryQuiz[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    apiFetch("/quiz/")
      .then((r) => r.json())
      .then((data: QuizDTO[]) => {
        const mapped: LibraryQuiz[] = (data ?? []).map((q) => ({
          id: q.id,
          title: q.title,
          category: q.category,
          difficulty: q.difficulty,
          tags: q.tags ?? [],
          coverImageUrl: q.coverImageUrl ?? null,
          archived: q.archived,
          questionCount: 0,
          totalPlays: 0,
          lastRun: null,
        }));
        setQuizzes(mapped);
        setLoading(false);
      })
      .catch(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div style={{ background: "#19191A", minHeight: "100vh", display: "flex", alignItems: "center", justifyContent: "center" }}>
        <span style={{ color: "#909499", fontFamily: "Inter, sans-serif", fontSize: 15 }}>Загружаем…</span>
      </div>
    );
  }

  return (
    <MyQuizzesView
      user={{ name: user?.name ?? "Вы", role: user?.role ?? "ORGANIZER" }}
      quizzes={quizzes}
    />
  );
}

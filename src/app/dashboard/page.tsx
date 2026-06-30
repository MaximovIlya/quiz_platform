"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/context/auth-context";
import { apiFetch } from "@/lib/api";
import OrganizerDashboard from "./_components/OrganizerDashboard";
import ParticipantDashboard from "./_components/ParticipantDashboard";
import type { LibraryQuiz } from "./_components/QuizLibrary";

type ActiveOrgSession = {
  sessionId: string;
  quizId: string;
  quizTitle: string;
  status: string;
};

type QuizFromAPI = {
  id: string;
  title: string;
  category: string;
  difficulty: string;
  tags: string[];
  coverImageUrl: string | null;
  archived: boolean;
  createdAt: string;
};

export default function DashboardPage() {
  const { user, loading } = useAuth();
  const router = useRouter();
  const [quizzes, setQuizzes] = useState<LibraryQuiz[]>([]);
  const [activeSession, setActiveSession] = useState<ActiveOrgSession | null>(null);
  const [dataLoading, setDataLoading] = useState(true);

  useEffect(() => {
    if (loading) return;
    if (!user) { router.replace("/login"); return; }

    if (user.role !== "ORGANIZER") {
      setDataLoading(false);
      return;
    }

    apiFetch("/quiz/")
      .then(r => r.json())
      .then((data: QuizFromAPI[]) => {
        const mapped: LibraryQuiz[] = data.map(q => ({
          id: q.id,
          title: q.title,
          category: q.category ?? "",
          questionCount: 0,
          totalPlays: 0,
          archived: q.archived ?? false,
          difficulty: q.difficulty ?? "",
          tags: q.tags ?? [],
          coverImageUrl: q.coverImageUrl ?? null,
          lastRun: null,
        }));
        setQuizzes(mapped);
      })
      .catch(() => {})
      .finally(() => setDataLoading(false));
  }, [user, loading, router]);

  if (loading || dataLoading) {
    return (
      <div style={{ minHeight: "100vh", background: "#19191A", display: "flex", alignItems: "center", justifyContent: "center" }}>
        <span style={{ color: "#909499", fontFamily: "Inter, sans-serif" }}>Загрузка…</span>
      </div>
    );
  }

  if (!user) return null;

  if (user.role === "ORGANIZER") {
    return (
      <OrganizerDashboard
        user={{ name: user.name, role: user.role }}
        stats={{
          totalQuizzes: quizzes.length,
          totalPlays: 0,
          avgScore: null,
          avgScoreDelta: null,
          activeRooms: 0,
        }}
        quizzes={quizzes}
        activeSession={activeSession}
      />
    );
  }

  return (
    <ParticipantDashboard
      user={{ name: user.name, role: user.role }}
      history={[]}
      activeSession={null}
    />
  );
}

package handler

import (
	"github.com/gofiber/fiber/v2"

	db "github.com/MaximovIlya/vk_practice_project/db/sqlc"
	"github.com/MaximovIlya/vk_practice_project/internal/repository"
)

type PlayHandler struct {
	store *repository.Store
}

func NewPlayHandler(store *repository.Store) *PlayHandler {
	return &PlayHandler{store: store}
}

type publicAnswer struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type publicQuestion struct {
	ID        string          `json:"id"`
	Text      string          `json:"text"`
	ImageUrl  *string         `json:"imageUrl"`
	Type      string          `json:"type"`
	Order     int32           `json:"order"`
	TimeLimit int32           `json:"timeLimit"`
	Points    int32           `json:"points"`
	Answers   []publicAnswer  `json:"answers"`
}

type playResponse struct {
	Session   sessionDTO       `json:"session"`
	Quiz      quizDTO          `json:"quiz"`
	Questions []publicQuestion `json:"questions"`
	HostName  string           `json:"hostName"`
}

// GET /play/:code — возвращает квиз без isCorrect для участника
func (h *PlayHandler) JoinByCode(c *fiber.Ctx) error {
	sess, err := h.store.GetSessionByRoomCode(c.Context(), c.Params("code"))
	if err != nil {
		return fiber.ErrNotFound
	}

	if sess.Status == db.SessionStatusFINISHED {
		return c.Status(fiber.StatusGone).JSON(fiber.Map{"error": "quiz already finished"})
	}

	quiz, err := h.store.GetQuizByID(c.Context(), sess.QuizID)
	if err != nil {
		return fiber.ErrNotFound
	}

	hostName := ""
	if author, err := h.store.GetUserByID(c.Context(), quiz.AuthorID); err == nil {
		hostName = author.Name
	}

	questions, err := h.store.GetQuestionsByQuizID(c.Context(), quiz.ID)
	if err != nil {
		return err
	}

	var pubQuestions []publicQuestion
	for _, q := range questions {
		answers, err := h.store.GetAnswersByQuestionID(c.Context(), q.ID)
		if err != nil {
			return err
		}

		var pubAnswers []publicAnswer
		for _, a := range answers {
			pubAnswers = append(pubAnswers, publicAnswer{ID: a.ID, Text: a.Text})
		}

		pubQuestions = append(pubQuestions, publicQuestion{
			ID:        q.ID,
			Text:      q.Text,
			ImageUrl:  q.ImageUrl,
			Type:      string(q.Type),
			Order:     q.Order,
			TimeLimit: q.TimeLimit,
			Points:    q.Points,
			Answers:   pubAnswers,
		})
	}

	return c.JSON(playResponse{
		Session:   mapSession(sess),
		Quiz:      mapQuiz(quiz),
		Questions: pubQuestions,
		HostName:  hostName,
	})
}

type resultsResponse struct {
	SessionID   string      `json:"sessionId"`
	QuizID      string      `json:"quizId"`
	QuizTitle   string      `json:"quizTitle"`
	HostName    string      `json:"hostName"`
	StartedAt   *string     `json:"startedAt"`
	PlayerCount int         `json:"playerCount"`
	Leaderboard []playerDTO `json:"leaderboard"`
}

type historyGameDTO struct {
	SessionID   string      `json:"sessionId"`
	StartedAt   string      `json:"startedAt"`
	PlayerCount int         `json:"playerCount"`
	WinnerName  *string     `json:"winnerName"`
	WinnerScore int32       `json:"winnerScore"`
}

type quizHistoryResponse struct {
	QuizID    string           `json:"quizId"`
	QuizTitle string           `json:"quizTitle"`
	Category  string           `json:"category"`
	Games     []historyGameDTO `json:"games"`
}

// GET /quiz/:id/history — история завершённых игр квиза
func (h *PlayHandler) GetQuizHistory(c *fiber.Ctx) error {
	quizID := c.Params("id")

	quiz, err := h.store.GetQuizByID(c.Context(), quizID)
	if err != nil {
		return fiber.ErrNotFound
	}

	sessions, err := h.store.GetFinishedSessionsByQuizID(c.Context(), quizID)
	if err != nil {
		sessions = nil
	}

	games := make([]historyGameDTO, 0, len(sessions))
	for _, s := range sessions {
		players, _ := h.store.GetLeaderboard(c.Context(), s.ID)
		var winnerName *string
		var winnerScore int32
		if len(players) > 0 {
			n := players[0].Name
			winnerName = &n
			winnerScore = players[0].Score
		}
		startedAt := ""
		if s.StartedAt.Valid {
			startedAt = s.StartedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}
		games = append(games, historyGameDTO{
			SessionID:   s.ID,
			StartedAt:   startedAt,
			PlayerCount: len(players),
			WinnerName:  winnerName,
			WinnerScore: winnerScore,
		})
	}

	return c.JSON(quizHistoryResponse{
		QuizID:    quiz.ID,
		QuizTitle: quiz.Title,
		Category:  string(quiz.Category),
		Games:     games,
	})
}

// GET /results/:sessionId — итоговый лидерборд
func (h *PlayHandler) GetResults(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	sess, err := h.store.GetSessionByID(c.Context(), sessionID)
	if err != nil {
		return fiber.ErrNotFound
	}

	quiz, _ := h.store.GetQuizByID(c.Context(), sess.QuizID)

	hostName := ""
	if author, err := h.store.GetUserByID(c.Context(), quiz.AuthorID); err == nil {
		hostName = author.Name
	}

	players, err := h.store.GetLeaderboard(c.Context(), sessionID)
	if err != nil {
		players = nil
	}

	result := make([]playerDTO, len(players))
	for i, p := range players {
		result[i] = mapPlayer(p)
	}

	var startedAt *string
	if sess.StartedAt.Valid {
		t := sess.StartedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		startedAt = &t
	}

	return c.JSON(resultsResponse{
		SessionID:   sess.ID,
		QuizID:      quiz.ID,
		QuizTitle:   quiz.Title,
		HostName:    hostName,
		StartedAt:   startedAt,
		PlayerCount: len(result),
		Leaderboard: result,
	})
}

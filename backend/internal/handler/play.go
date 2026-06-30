package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/MaximovIlya/vk_practice_project/internal/repository"
	db "github.com/MaximovIlya/vk_practice_project/db/sqlc"
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
	ID        string         `json:"id"`
	Text      string         `json:"text"`
	ImageURL  *string        `json:"imageUrl"`
	Type      db.QuestionType `json:"type"`
	Order     int32          `json:"order"`
	TimeLimit int32          `json:"timeLimit"`
	Points    int32          `json:"points"`
	Answers   []publicAnswer `json:"answers"`
}

type playResponse struct {
	Session   db.QuizSession   `json:"session"`
	Quiz      db.Quiz          `json:"quiz"`
	Questions []publicQuestion `json:"questions"`
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
			ImageURL:  q.ImageUrl,
			Type:      q.Type,
			Order:     q.Order,
			TimeLimit: q.TimeLimit,
			Points:    q.Points,
			Answers:   pubAnswers,
		})
	}

	return c.JSON(playResponse{
		Session:   sess,
		Quiz:      quiz,
		Questions: pubQuestions,
	})
}

// GET /results/:sessionId — итоговый лидерборд
func (h *PlayHandler) GetResults(c *fiber.Ctx) error {
	players, err := h.store.GetLeaderboard(c.Context(), c.Params("sessionId"))
	if err != nil {
		return fiber.ErrNotFound
	}
	return c.JSON(players)
}

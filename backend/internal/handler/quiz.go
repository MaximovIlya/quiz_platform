package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/MaximovIlya/vk_practice_project/internal/middleware"
	"github.com/MaximovIlya/vk_practice_project/internal/service"
	db "github.com/MaximovIlya/vk_practice_project/db/sqlc"
)

type QuizHandler struct {
	svc *service.QuizService
}

func NewQuizHandler(svc *service.QuizService) *QuizHandler {
	return &QuizHandler{svc: svc}
}

func (h *QuizHandler) Create(c *fiber.Ctx) error {
	var body struct {
		Title             string   `json:"title"`
		Description       string   `json:"description"`
		Category          string   `json:"category"`
		TimePerQuestion   int32    `json:"timePerQuestion"`
		PointsPerQuestion int32    `json:"pointsPerQuestion"`
		Scoring           string   `json:"scoring"`
		Difficulty        string   `json:"difficulty"`
		Tags              []string `json:"tags"`
		CoverImageURL     *string  `json:"coverImageUrl"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.ErrBadRequest
	}

	userID := c.Locals(middleware.CtxUserID).(string)

	quiz, err := h.svc.Create(c.Context(), service.CreateQuizParams{
		Title:             body.Title,
		Description:       body.Description,
		Category:          body.Category,
		TimePerQuestion:   body.TimePerQuestion,
		PointsPerQuestion: body.PointsPerQuestion,
		Scoring:           body.Scoring,
		Difficulty:        body.Difficulty,
		Tags:              body.Tags,
		CoverImageURL:     body.CoverImageURL,
		AuthorID:          userID,
	})
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(mapQuiz(quiz))
}

func (h *QuizHandler) GetByID(c *fiber.Ctx) error {
	userID := c.Locals(middleware.CtxUserID).(string)

	quiz, err := h.svc.GetByID(c.Context(), c.Params("id"), userID)
	if err != nil {
		return toHTTPError(err)
	}

	return c.JSON(mapQuizWithQuestions(quiz))
}

func (h *QuizHandler) ListMine(c *fiber.Ctx) error {
	userID := c.Locals(middleware.CtxUserID).(string)

	quizzes, err := h.svc.ListByAuthor(c.Context(), userID)
	if err != nil {
		return err
	}

	result := make([]quizDTO, len(quizzes))
	for i, q := range quizzes {
		result[i] = mapQuiz(q)
	}
	return c.JSON(result)
}

func (h *QuizHandler) Update(c *fiber.Ctx) error {
	var body db.UpdateQuizParams
	if err := c.BodyParser(&body); err != nil {
		return fiber.ErrBadRequest
	}

	userID := c.Locals(middleware.CtxUserID).(string)

	quiz, err := h.svc.Update(c.Context(), c.Params("id"), userID, body)
	if err != nil {
		return toHTTPError(err)
	}

	return c.JSON(mapQuiz(quiz))
}

func (h *QuizHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals(middleware.CtxUserID).(string)

	if err := h.svc.Delete(c.Context(), c.Params("id"), userID); err != nil {
		return toHTTPError(err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *QuizHandler) CreateQuestion(c *fiber.Ctx) error {
	var body struct {
		Text      string                   `json:"text"`
		ImageURL  *string                  `json:"imageUrl"`
		Type      db.QuestionType          `json:"type"`
		Order     int32                    `json:"order"`
		Tags      []string                 `json:"tags"`
		TimeLimit int32                    `json:"timeLimit"`
		Points    int32                    `json:"points"`
		Answers   []service.CreateAnswerParams `json:"answers"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.ErrBadRequest
	}

	userID := c.Locals(middleware.CtxUserID).(string)

	q, err := h.svc.CreateQuestion(c.Context(), userID, service.CreateQuestionParams{
		QuizID:    c.Params("id"),
		Text:      body.Text,
		ImageURL:  body.ImageURL,
		Type:      body.Type,
		Order:     body.Order,
		Tags:      body.Tags,
		TimeLimit: body.TimeLimit,
		Points:    body.Points,
		Answers:   body.Answers,
	})
	if err != nil {
		return toHTTPError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(mapQuestion(q.Question, q.Answers))
}

func (h *QuizHandler) UpdateQuestion(c *fiber.Ctx) error {
	var body struct {
		db.UpdateQuestionParams
		Answers []service.CreateAnswerParams `json:"answers"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.ErrBadRequest
	}

	userID := c.Locals(middleware.CtxUserID).(string)

	q, err := h.svc.UpdateQuestion(c.Context(), c.Params("questionId"), userID, body.UpdateQuestionParams, body.Answers)
	if err != nil {
		return toHTTPError(err)
	}

	return c.JSON(mapQuestion(q.Question, q.Answers))
}

func (h *QuizHandler) DeleteQuestion(c *fiber.Ctx) error {
	userID := c.Locals(middleware.CtxUserID).(string)

	if err := h.svc.DeleteQuestion(c.Context(), c.Params("questionId"), userID); err != nil {
		return toHTTPError(err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

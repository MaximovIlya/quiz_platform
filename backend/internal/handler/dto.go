package handler

import (
	"time"

	db "github.com/MaximovIlya/vk_practice_project/db/sqlc"
	"github.com/MaximovIlya/vk_practice_project/internal/service"
)

type sessionDTO struct {
	ID        string  `json:"id"`
	QuizID    string  `json:"quizId"`
	RoomCode  string  `json:"roomCode"`
	Status    string  `json:"status"`
	StartedAt *string `json:"startedAt"`
	CreatedAt string  `json:"createdAt"`
}

type playerDTO struct {
	SessionPlayerID string `json:"sessionPlayerID"`
	Score           int32  `json:"score"`
	UserID          string `json:"userId"`
	Name            string `json:"name"`
}

type sessionWithPlayersDTO struct {
	sessionDTO
	Players []playerDTO `json:"players"`
}

type quizDTO struct {
	ID                string   `json:"id"`
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	Category          string   `json:"category"`
	TimePerQuestion   int32    `json:"timePerQuestion"`
	PointsPerQuestion int32    `json:"pointsPerQuestion"`
	Scoring           string   `json:"scoring"`
	Difficulty        string   `json:"difficulty"`
	Tags              []string `json:"tags"`
	CoverImageUrl     *string  `json:"coverImageUrl"`
	Archived          bool     `json:"archived"`
	AuthorID          string   `json:"authorId"`
	CreatedAt         string   `json:"createdAt"`
}

type answerDTO struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	IsCorrect bool   `json:"isCorrect"`
}

type questionDTO struct {
	ID        string      `json:"id"`
	QuizID    string      `json:"quizId"`
	Text      string      `json:"text"`
	ImageUrl  *string     `json:"imageUrl"`
	Type      string      `json:"type"`
	Order     int32       `json:"order"`
	TimeLimit int32       `json:"timeLimit"`
	Points    int32       `json:"points"`
	Answers   []answerDTO `json:"answers"`
}

type quizWithQuestionsDTO struct {
	quizDTO
	Questions []questionDTO `json:"questions"`
}

func mapSession(s db.QuizSession) sessionDTO {
	dto := sessionDTO{
		ID:       s.ID,
		QuizID:   s.QuizID,
		RoomCode: s.RoomCode,
		Status:   string(s.Status),
		CreatedAt: s.CreatedAt.Time.Format(time.RFC3339),
	}
	if s.StartedAt.Valid {
		t := s.StartedAt.Time.Format(time.RFC3339)
		dto.StartedAt = &t
	}
	return dto
}

func mapPlayer(p db.GetLeaderboardRow) playerDTO {
	return playerDTO{
		SessionPlayerID: p.SessionPlayerID,
		Score:           p.Score,
		UserID:          p.UserID,
		Name:            p.Name,
	}
}

func mapSessionWithPlayers(s db.QuizSession, players []db.GetLeaderboardRow) sessionWithPlayersDTO {
	mapped := make([]playerDTO, len(players))
	for i, p := range players {
		mapped[i] = mapPlayer(p)
	}
	return sessionWithPlayersDTO{
		sessionDTO: mapSession(s),
		Players:    mapped,
	}
}

func mapQuiz(q db.Quiz) quizDTO {
	return quizDTO{
		ID:                q.ID,
		Title:             q.Title,
		Description:       q.Description,
		Category:          q.Category,
		TimePerQuestion:   q.TimePerQuestion,
		PointsPerQuestion: q.PointsPerQuestion,
		Scoring:           q.Scoring,
		Difficulty:        q.Difficulty,
		Tags:              q.Tags,
		CoverImageUrl:     q.CoverImageUrl,
		Archived:          q.Archived,
		AuthorID:          q.AuthorID,
		CreatedAt:         q.CreatedAt.Time.Format(time.RFC3339),
	}
}

func mapAnswer(a db.Answer) answerDTO {
	return answerDTO{
		ID:        a.ID,
		Text:      a.Text,
		IsCorrect: a.IsCorrect,
	}
}

func mapQuizWithQuestions(qwq *service.QuizWithQuestions) quizWithQuestionsDTO {
	questions := make([]questionDTO, len(qwq.Questions))
	for i, q := range qwq.Questions {
		questions[i] = mapQuestion(q.Question, q.Answers)
	}
	return quizWithQuestionsDTO{
		quizDTO:   mapQuiz(qwq.Quiz),
		Questions: questions,
	}
}

func mapQuestion(q db.Question, answers []db.Answer) questionDTO {
	mapped := make([]answerDTO, len(answers))
	for i, a := range answers {
		mapped[i] = mapAnswer(a)
	}
	return questionDTO{
		ID:        q.ID,
		QuizID:    q.QuizID,
		Text:      q.Text,
		ImageUrl:  q.ImageUrl,
		Type:      string(q.Type),
		Order:     q.Order,
		TimeLimit: q.TimeLimit,
		Points:    q.Points,
		Answers:   mapped,
	}
}

package service

import (
	"context"
	"errors"

	"github.com/oklog/ulid/v2"

	"github.com/MaximovIlya/vk_practice_project/internal/repository"
	db "github.com/MaximovIlya/vk_practice_project/db/sqlc"
)

var ErrNotFound = errors.New("not found")
var ErrForbidden = errors.New("forbidden")

type QuizService struct {
	store *repository.Store
}

func NewQuizService(store *repository.Store) *QuizService {
	return &QuizService{store: store}
}

type CreateQuizParams struct {
	Title              string
	Description        string
	Category           string
	TimePerQuestion    int32
	PointsPerQuestion  int32
	Scoring            string
	Difficulty         string
	Tags               []string
	CoverImageURL      *string
	AuthorID           string
}

func (s *QuizService) Create(ctx context.Context, p CreateQuizParams) (db.Quiz, error) {
	return s.store.CreateQuiz(ctx, db.CreateQuizParams{
		ID:                 ulid.Make().String(),
		Title:              p.Title,
		Description:        p.Description,
		Category:           p.Category,
		TimePerQuestion:    p.TimePerQuestion,
		PointsPerQuestion:  p.PointsPerQuestion,
		Scoring:            p.Scoring,
		Difficulty:         p.Difficulty,
		Tags:               p.Tags,
		CoverImageUrl:      p.CoverImageURL,
		AuthorID:           p.AuthorID,
	})
}

type QuizWithQuestions struct {
	db.Quiz
	Questions []QuestionWithAnswers `json:"questions"`
}

type QuestionWithAnswers struct {
	db.Question
	Answers []db.Answer `json:"answers"`
}

func (s *QuizService) GetByID(ctx context.Context, quizID, userID string) (*QuizWithQuestions, error) {
	quiz, err := s.store.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, ErrNotFound
	}
	if quiz.AuthorID != userID {
		return nil, ErrForbidden
	}

	questions, err := s.store.GetQuestionsByQuizID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	result := &QuizWithQuestions{Quiz: quiz}
	for _, q := range questions {
		answers, err := s.store.GetAnswersByQuestionID(ctx, q.ID)
		if err != nil {
			return nil, err
		}
		result.Questions = append(result.Questions, QuestionWithAnswers{
			Question: q,
			Answers:  answers,
		})
	}

	return result, nil
}

func (s *QuizService) ListByAuthor(ctx context.Context, authorID string) ([]db.Quiz, error) {
	return s.store.ListQuizzesByAuthor(ctx, authorID)
}

func (s *QuizService) Update(ctx context.Context, quizID, userID string, p db.UpdateQuizParams) (db.Quiz, error) {
	quiz, err := s.store.GetQuizByID(ctx, quizID)
	if err != nil {
		return db.Quiz{}, ErrNotFound
	}
	if quiz.AuthorID != userID {
		return db.Quiz{}, ErrForbidden
	}
	p.ID = quizID
	return s.store.UpdateQuiz(ctx, p)
}

func (s *QuizService) Delete(ctx context.Context, quizID, userID string) error {
	quiz, err := s.store.GetQuizByID(ctx, quizID)
	if err != nil {
		return ErrNotFound
	}
	if quiz.AuthorID != userID {
		return ErrForbidden
	}
	return s.store.DeleteQuiz(ctx, quizID)
}

// Questions

type CreateQuestionParams struct {
	QuizID    string
	Text      string
	ImageURL  *string
	Type      db.QuestionType
	Order     int32
	Tags      []string
	TimeLimit int32
	Points    int32
	Answers   []CreateAnswerParams
}

type CreateAnswerParams struct {
	Text      string
	IsCorrect bool
}

func (s *QuizService) CreateQuestion(ctx context.Context, userID string, p CreateQuestionParams) (QuestionWithAnswers, error) {
	quiz, err := s.store.GetQuizByID(ctx, p.QuizID)
	if err != nil {
		return QuestionWithAnswers{}, ErrNotFound
	}
	if quiz.AuthorID != userID {
		return QuestionWithAnswers{}, ErrForbidden
	}

	q, err := s.store.CreateQuestion(ctx, db.CreateQuestionParams{
		ID:        ulid.Make().String(),
		QuizID:    p.QuizID,
		Text:      p.Text,
		ImageUrl:  p.ImageURL,
		Type:      p.Type,
		Order:     p.Order,
		Tags:      p.Tags,
		TimeLimit: p.TimeLimit,
		Points:    p.Points,
	})
	if err != nil {
		return QuestionWithAnswers{}, err
	}

	result := QuestionWithAnswers{Question: q}
	for _, a := range p.Answers {
		ans, err := s.store.CreateAnswer(ctx, db.CreateAnswerParams{
			ID:         ulid.Make().String(),
			QuestionID: q.ID,
			Text:       a.Text,
			IsCorrect:  a.IsCorrect,
		})
		if err != nil {
			return QuestionWithAnswers{}, err
		}
		result.Answers = append(result.Answers, ans)
	}

	return result, nil
}

func (s *QuizService) UpdateQuestion(ctx context.Context, questionID, userID string, p db.UpdateQuestionParams, answers []CreateAnswerParams) (QuestionWithAnswers, error) {
	q, err := s.store.GetQuestionByID(ctx, questionID)
	if err != nil {
		return QuestionWithAnswers{}, ErrNotFound
	}

	quiz, err := s.store.GetQuizByID(ctx, q.QuizID)
	if err != nil || quiz.AuthorID != userID {
		return QuestionWithAnswers{}, ErrForbidden
	}

	p.ID = questionID
	updated, err := s.store.UpdateQuestion(ctx, p)
	if err != nil {
		return QuestionWithAnswers{}, err
	}

	if answers != nil {
		_ = s.store.DeleteAnswersByQuestionID(ctx, questionID)
		result := QuestionWithAnswers{Question: updated}
		for _, a := range answers {
			ans, err := s.store.CreateAnswer(ctx, db.CreateAnswerParams{
				ID:         ulid.Make().String(),
				QuestionID: questionID,
				Text:       a.Text,
				IsCorrect:  a.IsCorrect,
			})
			if err != nil {
				return QuestionWithAnswers{}, err
			}
			result.Answers = append(result.Answers, ans)
		}
		return result, nil
	}

	existing, _ := s.store.GetAnswersByQuestionID(ctx, questionID)
	return QuestionWithAnswers{Question: updated, Answers: existing}, nil
}

func (s *QuizService) DeleteQuestion(ctx context.Context, questionID, userID string) error {
	q, err := s.store.GetQuestionByID(ctx, questionID)
	if err != nil {
		return ErrNotFound
	}
	quiz, err := s.store.GetQuizByID(ctx, q.QuizID)
	if err != nil || quiz.AuthorID != userID {
		return ErrForbidden
	}
	return s.store.DeleteQuestion(ctx, questionID)
}

package service

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"strings"

	"github.com/oklog/ulid/v2"

	"github.com/MaximovIlya/vk_practice_project/internal/repository"
	db "github.com/MaximovIlya/vk_practice_project/db/sqlc"
)

type SessionService struct {
	store *repository.Store
}

func NewSessionService(store *repository.Store) *SessionService {
	return &SessionService{store: store}
}

type SessionWithPlayers struct {
	db.QuizSession
	Players []db.GetLeaderboardRow `json:"players"`
}

// GetOrCreate возвращает активную WAITING-сессию или создаёт новую
func (s *SessionService) GetOrCreate(ctx context.Context, quizID, userID string) (*SessionWithPlayers, error) {
	quiz, err := s.store.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, ErrNotFound
	}
	if quiz.AuthorID != userID {
		return nil, ErrForbidden
	}

	existing, err := s.store.GetLatestSessionByQuizID(ctx, quizID)
	if err == nil && existing.Status == db.SessionStatusWAITING {
		return s.withPlayers(ctx, existing)
	}

	code, err := generateRoomCode()
	if err != nil {
		return nil, err
	}

	sess, err := s.store.CreateSession(ctx, db.CreateSessionParams{
		ID:       ulid.Make().String(),
		QuizID:   quizID,
		RoomCode: code,
	})
	if err != nil {
		return nil, err
	}

	return s.withPlayers(ctx, sess)
}

func (s *SessionService) GetLatest(ctx context.Context, quizID, userID string) (*SessionWithPlayers, error) {
	quiz, err := s.store.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, ErrNotFound
	}
	if quiz.AuthorID != userID {
		return nil, ErrForbidden
	}

	sess, err := s.store.GetLatestSessionByQuizID(ctx, quizID)
	if err != nil {
		return nil, ErrNotFound
	}

	return s.withPlayers(ctx, sess)
}

func (s *SessionService) Delete(ctx context.Context, quizID, userID string) error {
	quiz, err := s.store.GetQuizByID(ctx, quizID)
	if err != nil {
		return ErrNotFound
	}
	if quiz.AuthorID != userID {
		return ErrForbidden
	}

	sess, err := s.store.GetLatestSessionByQuizID(ctx, quizID)
	if err != nil {
		return ErrNotFound
	}

	return s.store.DeleteSession(ctx, sess.ID)
}

func (s *SessionService) withPlayers(ctx context.Context, sess db.QuizSession) (*SessionWithPlayers, error) {
	players, err := s.store.GetLeaderboard(ctx, sess.ID)
	if err != nil {
		players = []db.GetLeaderboardRow{}
	}
	return &SessionWithPlayers{QuizSession: sess, Players: players}, nil
}

func generateRoomCode() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	code := base32.StdEncoding.EncodeToString(b)
	code = strings.TrimRight(code, "=")
	return code[:6], nil
}

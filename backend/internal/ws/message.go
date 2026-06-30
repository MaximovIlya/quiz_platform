package ws

import "encoding/json"

// InMessage — входящее сообщение от клиента
type InMessage struct {
	Type       string          `json:"type"`
	SessionID  string          `json:"sessionId"`
	RoomCode   string          `json:"roomCode"`
	QuestionID string          `json:"questionId"`
	AnswerIDs  []string        `json:"answerIds"`
	Name       string          `json:"name"`
}

// OutMessage — исходящее сообщение клиенту
type OutMessage struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

func encode(msgType string, payload any) []byte {
	b, _ := json.Marshal(OutMessage{Type: msgType, Payload: payload})
	return b
}

// Payload-структуры

type Player struct {
	UserID          string `json:"userId"`
	Name            string `json:"name"`
	Score           int32  `json:"score"`
	SessionPlayerID string `json:"sessionPlayerId"`
}

type QuestionStartedPayload struct {
	QuestionIndex int   `json:"questionIndex"`
	EndsAt        int64 `json:"endsAt"` // unix ms
}

type QuestionEndedPayload struct {
	QuestionIndex    int              `json:"questionIndex"`
	CorrectAnswerIDs []string         `json:"correctAnswerIds"`
	Votes            map[string]int64 `json:"votes"`
	IsLast           bool             `json:"isLast"`
}

type AnswerReceivedPayload struct {
	Votes         map[string]int64 `json:"votes"`
	TotalAnswered int64            `json:"totalAnswered"`
}

type AnswerResultPayload struct {
	Points        int32 `json:"points"`
	IsCorrect     bool  `json:"isCorrect"`
	PenaltyPoints int32 `json:"penaltyPoints,omitempty"`
}

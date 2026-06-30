package ws

import (
	"context"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/MaximovIlya/vk_practice_project/internal/repository"
	db "github.com/MaximovIlya/vk_practice_project/db/sqlc"
)

type clientMessage struct {
	client *Client
	msg    InMessage
}

// sessionState — in-memory состояние активной сессии
type sessionState struct {
	index     int
	endsAt    int64 // unix ms, 0 если вопрос не идёт
	reveal    *revealState
	qTimer    *time.Timer
	advTimer  *time.Timer
}

type revealState struct {
	correctAnswerIDs []string
	votes            map[string]int64
}

type Hub struct {
	store      *repository.Store
	rooms      map[string]map[*Client]bool // sessionID → clients
	sessions   map[string]*sessionState    // sessionID → state
	socketInfo map[*Client]struct{ sessionID, userID string }

	register   chan *Client
	unregister chan *Client
	incoming   chan clientMessage
}

func NewHub(store *repository.Store) *Hub {
	return &Hub{
		store:      store,
		rooms:      make(map[string]map[*Client]bool),
		sessions:   make(map[string]*sessionState),
		socketInfo: make(map[*Client]struct{ sessionID, userID string }),
		register:   make(chan *Client, 16),
		unregister: make(chan *Client, 16),
		incoming:   make(chan clientMessage, 256),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			// ничего не делаем при подключении — клиент должен прислать join-room
			_ = c

		case c := <-h.unregister:
			info, ok := h.socketInfo[c]
			if ok {
				h.broadcast(info.sessionID, "player-left", info.userID, c)
				h.removeFromRoom(info.sessionID, c)
				delete(h.socketInfo, c)
			}
			close(c.send)

		case cm := <-h.incoming:
			if cm.msg.Type == "_reveal" || cm.msg.Type == "_advance" {
				h.handleInternalEvent(cm.msg)
			} else {
				h.handleMessage(cm.client, cm.msg)
			}
		}
	}
}

func (h *Hub) handleMessage(c *Client, msg InMessage) {
	ctx := context.Background()

	switch msg.Type {
	case "organizer-join":
		h.onOrganizerJoin(ctx, c, msg.SessionID)
	case "join-room":
		h.onJoinRoom(ctx, c, msg.RoomCode, msg.Name)
	case "start-quiz":
		h.onStartQuiz(ctx, c, msg.SessionID)
	case "next-question":
		h.onNextQuestion(ctx, c, msg.SessionID)
	case "end-question":
		h.onEndQuestion(ctx, c, msg.SessionID)
	case "cancel-session":
		h.onCancelSession(ctx, c, msg.SessionID)
	case "submit-answer":
		h.onSubmitAnswer(ctx, c, msg.SessionID, msg.QuestionID, msg.AnswerIDs)
	}
}

// --- event handlers ---

func (h *Hub) onOrganizerJoin(ctx context.Context, c *Client, sessionID string) {
	h.addToRoom(sessionID, c)
	c.sessionID = sessionID

	sess, err := h.store.GetSessionByID(ctx, sessionID)
	if err != nil || sess.Status != db.SessionStatusACTIVE {
		return
	}

	state := h.sessions[sessionID]
	if state == nil {
		return
	}

	if state.reveal != nil {
		c.emit("question-ended", QuestionEndedPayload{
			QuestionIndex:    state.index,
			CorrectAnswerIDs: state.reveal.correctAnswerIDs,
			Votes:            state.reveal.votes,
		})
		players, _ := h.getLeaderboard(ctx, sessionID)
		c.emit("score-update", players)
	} else if state.endsAt > 0 {
		c.emit("question-started", QuestionStartedPayload{
			QuestionIndex: state.index,
			EndsAt:        state.endsAt,
		})
	}
}

func (h *Hub) onJoinRoom(ctx context.Context, c *Client, roomCode, name string) {
	sess, err := h.store.GetSessionByRoomCode(ctx, roomCode)
	if err != nil {
		c.emit("error", "room not found")
		return
	}
	if sess.Status == db.SessionStatusFINISHED {
		c.emit("error", "quiz already finished")
		return
	}

	sp, err := h.store.GetSessionPlayerByUserAndSession(ctx, db.GetSessionPlayerByUserAndSessionParams{
		SessionID: sess.ID,
		UserID:    c.userID,
	})
	if err != nil {
		sp, err = h.store.CreateSessionPlayer(ctx, db.CreateSessionPlayerParams{
			ID:        ulid.Make().String(),
			SessionID: sess.ID,
			UserID:    c.userID,
		})
		if err != nil {
			c.emit("error", "could not join session")
			return
		}
	}

	h.addToRoom(sess.ID, c)
	c.sessionID = sess.ID
	h.socketInfo[c] = struct{ sessionID, userID string }{sess.ID, c.userID}

	h.broadcast(sess.ID, "player-joined", Player{
		UserID:          c.userID,
		Name:            name,
		Score:           sp.Score,
		SessionPlayerID: sp.ID,
	}, nil)

	players, _ := h.getLeaderboard(ctx, sess.ID)
	c.emit("score-update", players)

	// восстановить состояние если сессия уже активна
	if sess.Status == db.SessionStatusACTIVE {
		state := h.sessions[sess.ID]
		if state != nil {
			if state.reveal != nil {
				c.emit("question-ended", QuestionEndedPayload{
					QuestionIndex:    state.index,
					CorrectAnswerIDs: state.reveal.correctAnswerIDs,
					Votes:            state.reveal.votes,
				})
			} else if state.endsAt > 0 {
				c.emit("question-started", QuestionStartedPayload{
					QuestionIndex: state.index,
					EndsAt:        state.endsAt,
				})
			}
		}
	}
}

func (h *Hub) onStartQuiz(ctx context.Context, c *Client, sessionID string) {
	if _, err := h.store.StartSession(ctx, sessionID); err != nil {
		return
	}
	h.broadcastAll(sessionID, "quiz-started", map[string]int{"questionIndex": 0})
	h.startQuestion(ctx, sessionID, 0)
}

func (h *Hub) onNextQuestion(ctx context.Context, c *Client, sessionID string) {
	state := h.sessions[sessionID]
	if state == nil {
		return
	}
	nextIdx := state.index + 1

	questions, err := h.getQuestions(ctx, sessionID)
	if err != nil {
		return
	}
	if nextIdx >= len(questions) {
		h.finishQuiz(ctx, sessionID)
		return
	}
	h.startQuestion(ctx, sessionID, nextIdx)
}

func (h *Hub) onEndQuestion(ctx context.Context, c *Client, sessionID string) {
	state := h.sessions[sessionID]
	if state != nil && state.qTimer != nil {
		state.qTimer.Stop()
		state.qTimer = nil
	}
	h.revealQuestion(ctx, sessionID)
}

func (h *Hub) onCancelSession(ctx context.Context, c *Client, sessionID string) {
	h.clearTimers(sessionID)
	delete(h.sessions, sessionID)
	_ = h.store.FinishSession(ctx, sessionID)
	h.broadcastAll(sessionID, "session-cancelled", nil)
}

func (h *Hub) onSubmitAnswer(ctx context.Context, c *Client, sessionID, questionID string, answerIDs []string) {
	sess, err := h.store.GetSessionByID(ctx, sessionID)
	if err != nil || sess.Status != db.SessionStatusACTIVE {
		return
	}

	sp, err := h.store.GetSessionPlayerByUserAndSession(ctx, db.GetSessionPlayerByUserAndSessionParams{
		SessionID: sessionID,
		UserID:    c.userID,
	})
	if err != nil {
		return
	}

	question, err := h.store.GetQuestionByID(ctx, questionID)
	if err != nil {
		return
	}

	allAnswers, err := h.store.GetAnswersByQuestionID(ctx, questionID)
	if err != nil {
		return
	}

	correctIDs := make(map[string]bool)
	for _, a := range allAnswers {
		if a.IsCorrect {
			correctIDs[a.ID] = true
		}
	}

	var isCorrect bool
	var earnedPoints, penaltyPoints int32

	if question.Type == db.QuestionTypeMULTIPLE {
		correctlySelected := int32(0)
		wronglySelected := int32(0)
		for _, id := range answerIDs {
			if correctIDs[id] {
				correctlySelected++
			} else {
				wronglySelected++
			}
		}
		isCorrect = correctlySelected == int32(len(correctIDs)) && wronglySelected == 0
		if len(correctIDs) > 0 {
			pointsPerCorrect := float64(question.Points) / float64(len(correctIDs))
			earned := float64(correctlySelected) * pointsPerCorrect
			penalty := float64(wronglySelected) * 0.5 * pointsPerCorrect
			earnedPoints = int32(max(0, earned-penalty))
			if wronglySelected > 0 {
				penaltyPoints = int32(penalty)
			}
		}
	} else {
		isCorrect = len(answerIDs) == len(correctIDs)
		if isCorrect {
			for _, id := range answerIDs {
				if !correctIDs[id] {
					isCorrect = false
					break
				}
			}
		}
		if isCorrect {
			earnedPoints = question.Points
		}
	}

	// scoring модификаторы
	if earnedPoints > 0 {
		quizID := sess.QuizID
		quiz, err := h.store.GetQuizByID(ctx, quizID)
		if err == nil {
			state := h.sessions[sessionID]
			switch quiz.Scoring {
			case "speed":
				if state != nil && state.endsAt > 0 {
					totalMs := int64(question.TimeLimit) * 1000
					remainingMs := max(0, state.endsAt-time.Now().UnixMilli())
					if totalMs > 0 {
						earnedPoints = int32(float64(earnedPoints) * float64(remainingMs) / float64(totalMs))
					}
				}
			case "streak":
				prior, _ := h.store.GetPlayerAnswersBySessionPlayer(ctx, sp.ID)
				streak := int32(1)
				for i := len(prior) - 1; i >= 0; i-- {
					if prior[i].IsCorrect {
						streak++
					} else {
						break
					}
				}
				earnedPoints = int32(float64(earnedPoints) * (1 + float64(streak-1)*0.1))
			}
		}
	}

	paID := ulid.Make().String()
	_, err = h.store.CreatePlayerAnswer(ctx, db.CreatePlayerAnswerParams{
		ID:              paID,
		SessionPlayerID: sp.ID,
		QuestionID:      questionID,
		IsCorrect:       isCorrect,
		Points:          earnedPoints,
	})
	if err != nil {
		return
	}

	for _, aID := range answerIDs {
		_ = h.store.ConnectPlayerAnswerToAnswer(ctx, db.ConnectPlayerAnswerToAnswerParams{
			PlayerAnswerID: paID,
			AnswerID:       aID,
		})
	}

	if earnedPoints > 0 {
		_, _ = h.store.IncrementPlayerScore(ctx, db.IncrementPlayerScoreParams{
			ID:    sp.ID,
			Score: earnedPoints,
		})
	}

	c.emit("answer-result", AnswerResultPayload{
		Points:        earnedPoints,
		IsCorrect:     isCorrect,
		PenaltyPoints: penaltyPoints,
	})

	votes, _ := h.getVotes(ctx, sessionID, questionID)
	total, _ := h.store.CountAnsweredPlayers(ctx, db.CountAnsweredPlayersParams{
		QuestionID: questionID,
		SessionID:  sessionID,
	})
	h.broadcastAll(sessionID, "answer-received", AnswerReceivedPayload{
		Votes:         votes,
		TotalAnswered: total,
	})
}

// --- quiz flow ---

func (h *Hub) startQuestion(ctx context.Context, sessionID string, idx int) {
	questions, err := h.getQuestions(ctx, sessionID)
	if err != nil || idx >= len(questions) {
		h.finishQuiz(ctx, sessionID)
		return
	}

	h.clearTimers(sessionID)

	q := questions[idx]
	endsAt := time.Now().UnixMilli() + int64(q.TimeLimit)*1000

	state := &sessionState{index: idx, endsAt: endsAt}
	h.sessions[sessionID] = state

	h.broadcastAll(sessionID, "question-started", QuestionStartedPayload{
		QuestionIndex: idx,
		EndsAt:        endsAt,
	})

	state.qTimer = time.AfterFunc(time.Duration(q.TimeLimit)*time.Second, func() {
		h.incoming <- clientMessage{msg: InMessage{Type: "_reveal", SessionID: sessionID}}
	})
}

func (h *Hub) revealQuestion(ctx context.Context, sessionID string) {
	state := h.sessions[sessionID]
	if state == nil {
		return
	}
	if state.qTimer != nil {
		state.qTimer.Stop()
		state.qTimer = nil
	}

	idx := state.index

	questions, err := h.getQuestions(ctx, sessionID)
	if err != nil || idx >= len(questions) {
		return
	}
	q := questions[idx]

	answers, _ := h.store.GetAnswersByQuestionID(ctx, q.ID)
	var correctIDs []string
	for _, a := range answers {
		if a.IsCorrect {
			correctIDs = append(correctIDs, a.ID)
		}
	}

	votes, _ := h.getVotes(ctx, sessionID, q.ID)
	isLast := idx >= len(questions)-1

	state.reveal = &revealState{correctAnswerIDs: correctIDs, votes: votes}

	h.broadcastAll(sessionID, "question-ended", QuestionEndedPayload{
		QuestionIndex:    idx,
		CorrectAnswerIDs: correctIDs,
		Votes:            votes,
		IsLast:           isLast,
	})

	players, _ := h.getLeaderboard(ctx, sessionID)
	h.broadcastAll(sessionID, "score-update", players)

	if isLast {
		return
	}

	nextIdx := idx + 1
	state.advTimer = time.AfterFunc(5*time.Second, func() {
		h.incoming <- clientMessage{msg: InMessage{Type: "_advance", SessionID: sessionID}}
	})
	_ = nextIdx
}

func (h *Hub) finishQuiz(ctx context.Context, sessionID string) {
	h.clearTimers(sessionID)
	delete(h.sessions, sessionID)
	_ = h.store.FinishSession(ctx, sessionID)
	players, _ := h.getLeaderboard(ctx, sessionID)
	h.broadcastAll(sessionID, "quiz-finished", players)
}

// handleMessage уже вызывается из Run(), добавляем внутренние события
func (h *Hub) handleInternalEvent(msg InMessage) {
	ctx := context.Background()
	switch msg.Type {
	case "_reveal":
		h.revealQuestion(ctx, msg.SessionID)
	case "_advance":
		state := h.sessions[msg.SessionID]
		if state == nil {
			return
		}
		questions, err := h.getQuestions(ctx, msg.SessionID)
		if err != nil {
			return
		}
		nextIdx := state.index + 1
		if nextIdx >= len(questions) {
			h.finishQuiz(ctx, msg.SessionID)
		} else {
			h.startQuestion(ctx, msg.SessionID, nextIdx)
		}
	}
}

// --- helpers ---

func (h *Hub) addToRoom(sessionID string, c *Client) {
	if h.rooms[sessionID] == nil {
		h.rooms[sessionID] = make(map[*Client]bool)
	}
	h.rooms[sessionID][c] = true
}

func (h *Hub) removeFromRoom(sessionID string, c *Client) {
	if room := h.rooms[sessionID]; room != nil {
		delete(room, c)
	}
}

func (h *Hub) broadcast(sessionID string, msgType string, payload any, except *Client) {
	data := encode(msgType, payload)
	for c := range h.rooms[sessionID] {
		if c == except {
			continue
		}
		select {
		case c.send <- data:
		default:
		}
	}
}

func (h *Hub) broadcastAll(sessionID string, msgType string, payload any) {
	h.broadcast(sessionID, msgType, payload, nil)
}

func (h *Hub) clearTimers(sessionID string) {
	if state := h.sessions[sessionID]; state != nil {
		if state.qTimer != nil {
			state.qTimer.Stop()
		}
		if state.advTimer != nil {
			state.advTimer.Stop()
		}
	}
}

func (h *Hub) getQuestions(ctx context.Context, sessionID string) ([]db.Question, error) {
	sess, err := h.store.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return h.store.GetQuestionsByQuizID(ctx, sess.QuizID)
}

func (h *Hub) getLeaderboard(ctx context.Context, sessionID string) ([]Player, error) {
	rows, err := h.store.GetLeaderboard(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	players := make([]Player, len(rows))
	for i, r := range rows {
		players[i] = Player{
			UserID:          r.UserID,
			Name:            r.Name,
			Score:           r.Score,
			SessionPlayerID: r.SessionPlayerID,
		}
	}
	return players, nil
}

func (h *Hub) getVotes(ctx context.Context, sessionID, questionID string) (map[string]int64, error) {
	rows, err := h.store.GetVotesForQuestion(ctx, db.GetVotesForQuestionParams{
		QuestionID: questionID,
		SessionID:  sessionID,
	})
	if err != nil {
		return nil, err
	}
	votes := make(map[string]int64, len(rows))
	for _, r := range rows {
		votes[r.AnswerID] = r.Votes
	}
	return votes, nil
}

func max[T int | int32 | int64 | float64](a, b T) T {
	if a > b {
		return a
	}
	return b
}

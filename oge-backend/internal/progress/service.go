package progress

import "context"

type Item struct {
	OgeNumber     int     `json:"oge_number"`
	SubtypeCode   string  `json:"subtype_code"`
	TopicTitle    string  `json:"topic_title,omitempty"`
	SubtopicTitle string  `json:"subtopic_title,omitempty"`
	TaskTitle     string  `json:"task_title,omitempty"`
	AttemptsCount int     `json:"attempts_count"`
	CorrectCount  int     `json:"correct_count"`
	MasteryScore  float64 `json:"mastery_score"`
}

type Stats struct {
	TotalAttempts  int     `json:"total_attempts"`
	CorrectCount   int     `json:"correct_count"`
	Accuracy       float64 `json:"accuracy"`
	AverageMastery float64 `json:"average_mastery"`
}

type Recommendation struct {
	OgeNumber    int     `json:"oge_number"`
	SubtypeCode  string  `json:"subtype_code"`
	MasteryScore float64 `json:"mastery_score"`
	Message      string  `json:"message"`
}

type Repository interface {
	GetProgress(ctx context.Context, userID int64) ([]Item, error)
	GetProgressStats(ctx context.Context, userID int64) (Stats, error)
	GetRecommendations(ctx context.Context, userID int64) ([]Recommendation, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetProgress(ctx context.Context, userID int64) ([]Item, error) {
	return s.repo.GetProgress(ctx, userID)
}

func (s *Service) GetStats(ctx context.Context, userID int64) (Stats, error) {
	return s.repo.GetProgressStats(ctx, userID)
}

func (s *Service) GetRecommendations(ctx context.Context, userID int64) ([]Recommendation, error) {
	return s.repo.GetRecommendations(ctx, userID)
}

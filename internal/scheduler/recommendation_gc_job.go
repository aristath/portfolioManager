package scheduler

import (
	"time"

	"github.com/aristath/sentinel/internal/modules/planning"
	"github.com/rs/zerolog"
)

type RecommendationGCJob struct {
	recommendationRepo planning.RecommendationRepositoryInterface
	maxAge             time.Duration
	log                zerolog.Logger
}

func NewRecommendationGCJob(recommendationRepo planning.RecommendationRepositoryInterface, maxAge time.Duration, log zerolog.Logger) *RecommendationGCJob {
	return &RecommendationGCJob{
		recommendationRepo: recommendationRepo,
		maxAge:             maxAge,
		log:                log.With().Str("job", "recommendation_gc").Logger(),
	}
}

func (j *RecommendationGCJob) Run() error {
	count, err := j.recommendationRepo.DeleteOlderThan(j.maxAge)
	if err != nil {
		j.log.Error().Err(err).Msg("Failed to delete old recommendations")
		return err
	}

	if count > 0 {
		j.log.Info().Int("deleted_count", count).Dur("max_age", j.maxAge).Msg("Garbage collection completed")
	}

	return nil
}

func (j *RecommendationGCJob) Name() string {
	return "recommendation_gc"
}

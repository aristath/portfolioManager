package handlers

import (
	"testing"

	"github.com/aristath/sentinel/internal/modules/opportunities"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestRegisterRoutes(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.Disabled)

	tagFilter := &mockTagFilter{}
	securityRepo := &mockSecurityRepo{}
	service := opportunities.NewService(tagFilter, securityRepo, logger)

	handler := NewHandler(service, nil, nil, logger)

	router := chi.NewRouter()

	// Should not panic
	assert.NotPanics(t, func() {
		handler.RegisterRoutes(router)
	}, "RegisterRoutes should not panic")
}

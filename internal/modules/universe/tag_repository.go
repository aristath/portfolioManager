package universe

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// TagRepository handles tag database operations
type TagRepository struct {
	universeDB *sql.DB // universe.db - tags table
	log        zerolog.Logger
}

// NewTagRepository creates a new tag repository
func NewTagRepository(universeDB *sql.DB, log zerolog.Logger) *TagRepository {
	return &TagRepository{
		universeDB: universeDB,
		log:        log.With().Str("repo", "tag").Logger(),
	}
}

// GetByID returns a tag by ID
func (r *TagRepository) GetByID(id string) (*Tag, error) {
	query := "SELECT id, name, created_at, updated_at FROM tags WHERE id = ?"

	rows, err := r.universeDB.Query(query, strings.ToLower(strings.TrimSpace(id)))
	if err != nil {
		return nil, fmt.Errorf("failed to query tag by ID: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil // Tag not found
	}

	var tag Tag
	var createdAtUnix, updatedAtUnix sql.NullInt64
	err = rows.Scan(&tag.ID, &tag.Name, &createdAtUnix, &updatedAtUnix)
	if err != nil {
		return nil, fmt.Errorf("failed to scan tag: %w", err)
	}

	// Timestamps are stored but not exposed in Tag model (internal only)
	_ = createdAtUnix
	_ = updatedAtUnix

	return &tag, nil
}

// GetAll returns all tags ordered by ID
func (r *TagRepository) GetAll() ([]Tag, error) {
	query := "SELECT id, name, created_at, updated_at FROM tags ORDER BY id"

	rows, err := r.universeDB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all tags: %w", err)
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var tag Tag
		var createdAtUnix, updatedAtUnix sql.NullInt64
		err := rows.Scan(&tag.ID, &tag.Name, &createdAtUnix, &updatedAtUnix)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		// Timestamps are stored but not exposed in Tag model (internal only)
		_ = createdAtUnix
		_ = updatedAtUnix
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tags: %w", err)
	}

	return tags, nil
}

// CreateOrGet creates a tag if it doesn't exist, or returns existing tag
func (r *TagRepository) CreateOrGet(tag Tag) (*Tag, error) {
	// Normalize tag ID
	tag.ID = strings.ToLower(strings.TrimSpace(tag.ID))
	if tag.ID == "" {
		return nil, fmt.Errorf("tag ID cannot be empty")
	}

	// Check if tag exists
	existing, err := r.GetByID(tag.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if tag exists: %w", err)
	}
	if existing != nil {
		return existing, nil
	}

	// Create new tag
	now := time.Now().Unix()

	// If name is empty, generate default name from ID
	if tag.Name == "" {
		tag.Name = generateDefaultTagName(tag.ID)
	}

	query := `
		INSERT INTO tags (id, name, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`

	_, err = r.universeDB.Exec(query, tag.ID, tag.Name, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to insert tag: %w", err)
	}

	r.log.Info().Str("tag_id", tag.ID).Str("tag_name", tag.Name).Msg("Tag created")

	return &tag, nil
}

// EnsureTagsExist ensures all tag IDs exist, creating them with default names if missing
func (r *TagRepository) EnsureTagsExist(tagIDs []string) error {
	if len(tagIDs) == 0 {
		return nil
	}

	// Begin transaction
	tx, err := r.universeDB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().Unix()

	for _, tagID := range tagIDs {
		// Skip empty tag IDs
		tagID = strings.ToLower(strings.TrimSpace(tagID))
		if tagID == "" {
			continue
		}

		// Check if tag exists
		var exists bool
		err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM tags WHERE id = ?)", tagID).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check if tag exists: %w", err)
		}

		if !exists {
			// Create tag with default name
			defaultName := generateDefaultTagName(tagID)
			_, err = tx.Exec(`
				INSERT INTO tags (id, name, created_at, updated_at)
				VALUES (?, ?, ?, ?)
			`, tagID, defaultName, now, now)
			if err != nil {
				return fmt.Errorf("failed to insert tag %s: %w", tagID, err)
			}
			r.log.Debug().Str("tag_id", tagID).Str("tag_name", defaultName).Msg("Tag auto-created")
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// generateDefaultTagName converts a tag ID to a human-readable name
// Example: 'value-opportunity' -> 'Value Opportunity'
func generateDefaultTagName(tagID string) string {
	// Split by hyphens
	parts := strings.Split(tagID, "-")
	var words []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len(part) > 0 {
			// Capitalize first letter, lowercase rest
			word := strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
			words = append(words, word)
		}
	}
	return strings.Join(words, " ")
}

package universe

import (
	"database/sql"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func setupTagTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create tags table
	_, err = db.Exec(`
		CREATE TABLE tags (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)
	`)
	require.NoError(t, err)

	return db
}

func TestTagRepository_GetByID_Existing(t *testing.T) {
	// Setup
	db := setupTagTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewTagRepository(db, log)
	var _ TagRepositoryInterface = repo // Verify interface implementation

	// Insert test tag
	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO tags (id, name, created_at, updated_at)
		VALUES ('value-opportunity', 'Value Opportunity', ?, ?)
	`, now, now)
	require.NoError(t, err)

	// Execute
	tag, err := repo.GetByID("value-opportunity")

	// Assert
	assert.NoError(t, err)
	require.NotNil(t, tag)
	assert.Equal(t, "value-opportunity", tag.ID)
	assert.Equal(t, "Value Opportunity", tag.Name)
}

func TestTagRepository_GetByID_NonExistent(t *testing.T) {
	// Setup
	db := setupTagTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewTagRepository(db, log)
	var _ TagRepositoryInterface = repo // Verify interface implementation

	// Execute
	tag, err := repo.GetByID("non-existent")

	// Assert
	assert.NoError(t, err)
	assert.Nil(t, tag)
}

func TestTagRepository_GetAll_Empty(t *testing.T) {
	// Setup
	db := setupTagTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewTagRepository(db, log)
	var _ TagRepositoryInterface = repo // Verify interface implementation

	// Execute
	tags, err := repo.GetAll()

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, tags)
}

func TestTagRepository_GetAll_Multiple(t *testing.T) {
	// Setup
	db := setupTagTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewTagRepository(db, log)
	var _ TagRepositoryInterface = repo // Verify interface implementation

	// Insert test tags
	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO tags (id, name, created_at, updated_at)
		VALUES
			('value-opportunity', 'Value Opportunity', ?, ?),
			('volatile', 'Volatile', ?, ?),
			('stable', 'Stable', ?, ?)
	`, now, now, now, now, now, now)
	require.NoError(t, err)

	// Execute
	tags, err := repo.GetAll()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, tags, 3)

	// Verify tags are sorted by ID
	tagIDs := make([]string, len(tags))
	for i, tag := range tags {
		tagIDs[i] = tag.ID
	}
	assert.Equal(t, []string{"stable", "value-opportunity", "volatile"}, tagIDs)
}

func TestTagRepository_CreateOrGet_New(t *testing.T) {
	// Setup
	db := setupTagTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewTagRepository(db, log)
	var _ TagRepositoryInterface = repo // Verify interface implementation

	// Execute
	tag := Tag{
		ID:   "value-opportunity",
		Name: "Value Opportunity",
	}
	result, err := repo.CreateOrGet(tag)

	// Assert
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "value-opportunity", result.ID)
	assert.Equal(t, "Value Opportunity", result.Name)

	// Verify tag was created in database
	dbTag, err := repo.GetByID("value-opportunity")
	assert.NoError(t, err)
	require.NotNil(t, dbTag)
	assert.Equal(t, "value-opportunity", dbTag.ID)
	assert.Equal(t, "Value Opportunity", dbTag.Name)
}

func TestTagRepository_CreateOrGet_Existing(t *testing.T) {
	// Setup
	db := setupTagTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewTagRepository(db, log)
	var _ TagRepositoryInterface = repo // Verify interface implementation

	// Insert existing tag
	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO tags (id, name, created_at, updated_at)
		VALUES ('value-opportunity', 'Value Opportunity', ?, ?)
	`, now, now)
	require.NoError(t, err)

	// Execute
	tag := Tag{
		ID:   "value-opportunity",
		Name: "Value Opportunity",
	}
	result, err := repo.CreateOrGet(tag)

	// Assert
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "value-opportunity", result.ID)
	assert.Equal(t, "Value Opportunity", result.Name)
}

func TestTagRepository_EnsureTagsExist_NewTags(t *testing.T) {
	// Setup
	db := setupTagTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewTagRepository(db, log)
	var _ TagRepositoryInterface = repo // Verify interface implementation

	// Execute
	tagIDs := []string{"value-opportunity", "volatile", "stable"}
	err := repo.EnsureTagsExist(tagIDs)

	// Assert
	assert.NoError(t, err)

	// Verify all tags were created with default names
	expectedNames := map[string]string{
		"value-opportunity": "Value Opportunity",
		"volatile":          "Volatile",
		"stable":            "Stable",
	}
	for _, tagID := range tagIDs {
		tag, err := repo.GetByID(tagID)
		assert.NoError(t, err)
		require.NotNil(t, tag)
		assert.Equal(t, tagID, tag.ID)
		// Verify default name generation
		expectedName, ok := expectedNames[tagID]
		require.True(t, ok, "Expected name not defined for tag: %s", tagID)
		assert.Equal(t, expectedName, tag.Name)
	}
}

func TestTagRepository_EnsureTagsExist_ExistingTags(t *testing.T) {
	// Setup
	db := setupTagTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewTagRepository(db, log)
	var _ TagRepositoryInterface = repo // Verify interface implementation

	// Insert existing tag
	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO tags (id, name, created_at, updated_at)
		VALUES ('value-opportunity', 'Value Opportunity', ?, ?)
	`, now, now)
	require.NoError(t, err)

	// Execute
	tagIDs := []string{"value-opportunity", "volatile"}
	err = repo.EnsureTagsExist(tagIDs)

	// Assert
	assert.NoError(t, err)

	// Verify existing tag unchanged
	existingTag, err := repo.GetByID("value-opportunity")
	assert.NoError(t, err)
	require.NotNil(t, existingTag)
	assert.Equal(t, "Value Opportunity", existingTag.Name) // Original name preserved

	// Verify new tag created
	newTag, err := repo.GetByID("volatile")
	assert.NoError(t, err)
	require.NotNil(t, newTag)
	assert.Equal(t, "volatile", newTag.ID)
}

func TestTagRepository_EnsureTagsExist_EmptyArray(t *testing.T) {
	// Setup
	db := setupTagTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewTagRepository(db, log)
	var _ TagRepositoryInterface = repo // Verify interface implementation

	// Execute
	err := repo.EnsureTagsExist([]string{})

	// Assert
	assert.NoError(t, err)
}

func TestTagRepository_EnsureTagsExist_EmptyTagID(t *testing.T) {
	// Setup
	db := setupTagTestDB(t)
	defer db.Close()

	log := zerolog.New(nil).Level(zerolog.Disabled)
	repo := NewTagRepository(db, log)
	var _ TagRepositoryInterface = repo // Verify interface implementation

	// Execute - empty string should be skipped
	err := repo.EnsureTagsExist([]string{"value-opportunity", "", "volatile"})

	// Assert
	assert.NoError(t, err)

	// Verify only non-empty tags were created
	tag1, err := repo.GetByID("value-opportunity")
	assert.NoError(t, err)
	assert.NotNil(t, tag1)

	tag2, err := repo.GetByID("volatile")
	assert.NoError(t, err)
	assert.NotNil(t, tag2)

	emptyTag, err := repo.GetByID("")
	assert.NoError(t, err)
	assert.Nil(t, emptyTag)
}

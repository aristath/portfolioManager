package universe

// TagRepositoryInterface defines the contract for tag repository operations
type TagRepositoryInterface interface {
	// GetByID returns a tag by ID
	GetByID(id string) (*Tag, error)

	// GetAll returns all tags ordered by ID
	GetAll() ([]Tag, error)

	// CreateOrGet creates a tag if it doesn't exist, or returns existing tag
	CreateOrGet(tag Tag) (*Tag, error)

	// EnsureTagsExist ensures all tag IDs exist, creating them with default names if missing
	EnsureTagsExist(tagIDs []string) error
}

// Compile-time check that TagRepository implements TagRepositoryInterface
var _ TagRepositoryInterface = (*TagRepository)(nil)

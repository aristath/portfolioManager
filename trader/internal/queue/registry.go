package queue

// Handler processes a job
type Handler func(*Job) error

// Registry maps job types to handlers
type Registry struct {
	handlers map[JobType]Handler
}

// NewRegistry creates a new job registry
func NewRegistry() *Registry {
	return &Registry{
		handlers: make(map[JobType]Handler),
	}
}

// Register registers a handler for a job type
func (r *Registry) Register(jobType JobType, handler Handler) {
	r.handlers[jobType] = handler
}

// Get retrieves a handler for a job type
func (r *Registry) Get(jobType JobType) (Handler, bool) {
	handler, exists := r.handlers[jobType]
	return handler, exists
}

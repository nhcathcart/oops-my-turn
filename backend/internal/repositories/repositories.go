package repositories

import "database/sql"

// Repositories bundles the app's repository implementations.
type Repositories struct {
	User UserRepository
}

// NewRepositories constructs the repository bundle used by handlers.
func NewRepositories(db *sql.DB) Repositories {
	return Repositories{
		User: NewUserRepository(db),
	}
}

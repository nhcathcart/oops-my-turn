package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/opt/omit"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"go.jetify.com/typeid"

	models "github.com/nhcathcart/oops-my-turn/backend/models/generated"
)

type UserRepository interface {
	Upsert(ctx context.Context, googleID, email, firstName, lastName string) (*models.User, error)
}

type PostgresUserRepository struct {
	db bob.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &PostgresUserRepository{db: bob.NewDB(db)}
}

func (r *PostgresUserRepository) Upsert(ctx context.Context, googleID, email, firstName, lastName string) (*models.User, error) {
	userID, err := typeid.New[UserID]()
	if err != nil {
		return nil, fmt.Errorf("generate user id: %w", err)
	}

	setter := &models.UserSetter{
		ID:        omit.From(userID.String()),
		GoogleID:  omit.From(googleID),
		Email:     omit.From(email),
		FirstName: omit.From(firstName),
		LastName:  omit.From(lastName),
	}

	query := models.Users.Insert(
		setter,
		im.OnConflict("google_id").DoUpdate(
			im.SetExcluded("email", "first_name", "last_name"),
		),
	)

	user, err := query.One(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("upsert user: %w", err)
	}

	return user, nil
}

package migrate

import "github.com/google/uuid"

// TODO: Migrate User in resolver methods where needed. Make sure not to miss any endpoints!

func (m *Migrator) MigrateV1UserToV2(userID uuid.UUID) (bool, error) {
	userIsV2, err := userIsV2Model(userID)
	if err != nil {
		return false, err
	}

	if userIsV2 {
		return false, nil
	}

	// TODO: 1. Get v1user model

	// TODO: 2. Count number of v1Projects
	// TODO: Find out if we can query for v1Projects using the newer v2 query??

	// TODO: 3. Remove old v1user model from DB and add new v2 user model
	return true, nil
}

func userIsV2Model(userID uuid.UUID) (bool, error) {
	// TODO: Check if user in DB for userID is a v1User model.
	return false, nil
}

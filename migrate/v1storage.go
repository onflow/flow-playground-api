package migrate

import (
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func v1GetAccountsForProject(db *gorm.DB, pID uuid.UUID, accs *[]*v1Account) error {
	return db.
		Where(&model.Account{ProjectID: pID}).
		Order("\"index\" asc").
		Find(accs).
		Error
}

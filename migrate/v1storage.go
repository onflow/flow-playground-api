package migrate

import (
	"github.com/dapperlabs/flow-playground-api/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func v1_0_0GetAccountsForProject(db *gorm.DB, pID uuid.UUID, accs *[]*v1_0_0Account) error {
	return db.
		Where(&model.Account{ProjectID: pID}).
		Order("\"index\" asc").
		Find(accs).
		Error
}

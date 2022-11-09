package storage

import (
	"golang.org/x/xerrors"

	"github.com/filecoin-project/venus-auth/log"
)

var migrationSchedules = map[uint64]struct {
	from, to int64
	migrate  func(Store) error
}{
	0: {from: 0, to: 1, migrate: Store.MigrateToV1},
	1: {from: 1, to: 2, migrate: Store.MigrateToV2},
	2: {from: 2, to: 3, migrate: Store.MigrateToV3},
}

func StoreMigrate(store Store) error {
	for {
		v, err := store.Version()
		if err != nil {
			return err
		}
		mf, exists := migrationSchedules[v]
		if !exists {
			return nil
		}
		if err := mf.migrate(store); err != nil {
			return xerrors.Errorf("migrate from store version:%d failed:%w", v, err)
		}
		log.Infof("migrate from:%d, to:%d success.", mf.from, mf.to)
	}
}

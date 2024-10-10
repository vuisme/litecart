package app

import (
	"github.com/vuisme/litecart/internal/base"
	"github.com/vuisme/litecart/migrations"
	"github.com/vuisme/litecart/pkg/fsutil"
)

// Init is ...
func Init() error {
	dirsToCheck := []struct {
		path string
		name string
	}{
		{"./lc_uploads", "lc_uploads"},
		{"./lc_digitals", "lc_digitals"},
	}

	for _, dir := range dirsToCheck {
		if err := fsutil.MkDirs(0o775, dir.path); err != nil {
			log.Err(err).Send()
			return err
		}
	}

	if _, err := base.New("./lc_base/data.db", migrations.Embed()); err != nil {
		log.Err(err).Send()
		return err
	}

	return nil
}

// Migrate is ...
func Migrate() error {
	if err := base.Migrate("./lc_base/data.db", migrations.Embed()); err != nil {
		return err
	}

	return nil
}

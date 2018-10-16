package export

import "git.vlct.io/vaulter/vaulter/pkg/logger"

var log = logger.New()

type Exporter interface {
	Export(string) ([]byte, error)
}

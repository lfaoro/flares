package export

import "github.com/vlct-io/pkg/logger"

var log = logger.New()

// CloudDNS is a dns cloud provider.
type CloudDNS interface {
	ExportDNS(string) ([]byte, error)
}

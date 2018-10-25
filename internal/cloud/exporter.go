package cloud

import "github.com/vlct-io/pkg/logger"

var log = logger.New()

// Service is a dns cloud provider.
type Service interface {
	DNSTableFor(string) ([]byte, error)
}

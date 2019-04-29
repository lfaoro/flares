package cloud

// Service is a dns cloud provider.
type Service interface {
	DNSTableFor(string) ([]byte, error)
}

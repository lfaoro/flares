package vcs

type VCS interface {
	Clone()
	Add()
	// Remove()
	Push()
}

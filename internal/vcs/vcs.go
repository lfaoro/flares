package vcs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"git.vlct.io/vaulter/vaulter/pkg/svc"
)

type VCS interface {
	Clone()
	Add()
	// Remove()
	Push()
	Clean()
}

type Git struct {
	Repository string
	Username   string
	Password   string
	Directory  string
	repo       string
	name       string
	once       sync.Once
}

func (g *Git) init() {
	g.repo = g.assemble()
	g.name = svc.RandString(12)
	os.MkdirAll(g.Directory, 0755)
}

func (g Git) assemble() string {
	split := strings.SplitAfter(g.Repository, "//")
	return fmt.Sprintf("%v%v:%v@%v", split[0], g.Username, g.Password, split[1])
}

func (g *Git) Clone() error {
	g.once.Do(func() {
		g.init()
	})
	fmt.Println(g.Directory, g.repo, g.name)
	cmd := exec.Command("git", "clone", g.repo, g.name)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Dir = g.Directory
	return cmd.Run()
}

func (g *Git) Add(filePath string) error {
	g.once.Do(func() {
		g.init()
	})
	cmd := exec.Command("cp", "-frv", filePath, filepath.Join(g.Directory, g.name))
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("git", "add", "--all", "-v")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Dir = g.Directory + g.name
	if err := cmd.Run(); err != nil {
		return err
	}
	split := strings.Split(filePath, "/")
	fileName := fmt.Sprintf("Add %v", split[len(split)-1])
	cmd = exec.Command("git", "commit", "-am", fileName)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Dir = g.Directory + g.name
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (g *Git) Push() error {
	g.once.Do(func() {
		g.init()
	})
	cmd := exec.Command("git", "push", "--progress", "-v")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Dir = g.Directory + g.name
	return cmd.Run()
}

func (g *Git) Clean(dir string) {
	g.once.Do(func() {
		g.init()
	})
	cmd := exec.Command("rm", "-rf", g.Directory)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Run()
}

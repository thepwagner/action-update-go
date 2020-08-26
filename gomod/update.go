package gomod

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

func Update(pkgName, targetVersion string) error {
	if err := updateGoMod(pkgName, targetVersion); err != nil {
		return err
	}
	if err := updateGoSum(); err != nil {
		return err
	}

	if hasVendor() {
		if err := updateVendor(); err != nil {
			return err
		}
	}
	return nil
}

func updateGoMod(pkgName, targetVersion string) error {
	goMod, err := Parse()
	if err != nil {
		return err
	}

	// Update and write:
	if err := goMod.AddRequire(pkgName, targetVersion); err != nil {
		return fmt.Errorf("adding requirement: %w", err)
	}
	updated, err := goMod.Format()
	if err != nil {
		return fmt.Errorf("formatting go.mod: %w", err)
	}
	if err := ioutil.WriteFile(goModFn, updated, 0644); err != nil {
		return fmt.Errorf("writing go.mod: %w", err)
	}
	return nil
}

func updateGoSum() error {
	if err := exec.Command("go", "get", "-d", "-v").Run(); err != nil {
		return fmt.Errorf("updating go.sum: %w", err)
	}

	// TODO: toggle via environment? via reading config file?
	if err := exec.Command("go", "mod", "tidy").Run(); err != nil {
		return fmt.Errorf("tidying go.sum: %w", err)
	}
	return nil
}

func hasVendor() bool {
	_, err := os.Stat("vendor/modules.txt")
	return err == nil
}

func updateVendor() error {
	if err := exec.Command("go", "mod", "vendor").Run(); err != nil {
		return fmt.Errorf("updating go.sum")
	}
	return nil
}

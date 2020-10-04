package updater

// Dependency is an updatable external artifact
type Dependency struct {
	// Path is the name of the dependency.
	Path    string
	Version string
	// Indirect indicates this dependency is only required by a dependency of the target project.
	Indirect bool
}

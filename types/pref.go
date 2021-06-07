package types

type Options struct {
	Name          string
	From          string
	Executable    string
	Force         bool
	SelectDefault bool
	Integrate     bool
	DoNotFilter   bool
	FromGithub    bool
	Silent        bool
}

type InstallOptions struct {
	Name string
	Executable string
	From string
	FromGithub bool
	RemovePreviousVersions bool
	DoNotFilter   bool
	Silent bool
}


func (options InstallOptions) ToRemoveOptions() RemoveOptions {
	return RemoveOptions{Executable: options.Executable}
}

type RemoveOptions struct {
	Executable string
}

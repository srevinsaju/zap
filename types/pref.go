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
}

package model

// RepoURLComponents holds parsed GitHub URL components
type RepoURLComponents struct {
	Owner      string
	Repository string
	Ref        string
	Dir        string
	FilePath   string
	IsFile     bool
}

type FileInfo struct {
	Path string
	Size int64
	SHA  string
}

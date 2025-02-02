package structures

type FileInfo struct {
	Name         string
	UserAndGroup string
	Permissions  string
	Size         string
	RawSize      int64
	ModTime      string
	IsDir        bool
	Hidden       bool
	Path         string
	Selected     bool
	Color        string
}

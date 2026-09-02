package highlighter

import (
	"io/fs"
	"os"
)

// repoRoot returns the repository root as an fs.FS.
func repoRoot() fs.FS {
	return os.DirFS("../..")
}

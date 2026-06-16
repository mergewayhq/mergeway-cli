package data

import "github.com/mergewayhq/mergeway-cli/internal/fileutil"

func defaultFileOps() fileutil.Ops {
	return fileutil.OS
}

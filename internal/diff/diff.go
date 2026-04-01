package diff

type Options struct {
	Root   string
	Config string
	Args   []string
	JSON   bool
}

func Run(opts Options) (string, error) {
	snapshots, err := resolveDiffSnapshots(opts.Root, opts.Args)
	if err != nil {
		return "", err
	}

	corpora, err := loadDiffDataCorpora(opts.Root, opts.Config, snapshots.Left, snapshots.Right)
	if err != nil {
		return "", err
	}

	leftDB, err := buildLogicalDatabase(corpora.Left)
	if err != nil {
		return "", err
	}
	rightDB, err := buildLogicalDatabase(corpora.Right)
	if err != nil {
		return "", err
	}

	result, err := diffLogicalDatabases(leftDB, rightDB)
	if err != nil {
		return "", err
	}

	if opts.JSON {
		payload, err := marshalDiffResultJSON(result)
		if err != nil {
			return "", err
		}
		return string(payload) + "\n", nil
	}

	return renderDiffResult(result), nil
}

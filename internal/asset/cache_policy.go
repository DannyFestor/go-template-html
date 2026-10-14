package asset

type cachePolicy struct {
	current string
	stale   string
}

func fixedBuildCaching() cachePolicy {
	return cachePolicy{current: "public, max-age=31536000, immutable", stale: "no-cache"}
}

func liveBuildCaching() cachePolicy {
	return cachePolicy{current: "no-store", stale: "no-store"}
}

func (c cachePolicy) cacheControl(file resolved) string {
	if file.current {
		return c.current
	}

	return c.stale
}

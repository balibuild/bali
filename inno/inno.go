package inno

var (
	innoExePath string = "iscc"
	innoVersion string
)

func InnoVersion() string {
	return innoVersion
}

func InnoExePath() string {
	return innoExePath
}

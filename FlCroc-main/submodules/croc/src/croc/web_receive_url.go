package croc

// WebReceiveURL returns the web receive URL for a share code.
//
// The URL format belongs to croc (webReceiveURL in croc.go); embedders that need
// to display or encode it must ask for it instead of rebuilding the string, or
// they will keep serving a stale URL if the format changes.
func WebReceiveURL(code string) string {
	return webReceiveURL(code)
}

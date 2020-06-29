package main

/*
// Generates the random strings which are used as identifiers for each task
// They need to be large enough to make collisions of tasks not a concern
// Currently the key space is 7.95 * 10^24
func GenRandomIdentifier() string {
	// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
	b := ""
	rand.Seed(time.Now().UTC().UnixNano())
	for i := 0; i < DefaultIdentifierLength; i++ {
		b = b + string(AlphaNumChars[rand.Intn(len(AlphaNumChars))])
	}

	return b
}
*/

/*
// Check to see if a flag has been removed by the RemoveBrowserFlags setting
func IsRemoved(toRemove []string, candidate string) bool {
	for _, x := range toRemove {
		if candidate == x {
			return true
		}
	}

	return false
}

// DirNameFromURL takes a URL and sanitizes/escapes it so it can safely be used as a filename
func DirNameFromURL(s string) (string, error) {
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return "", err
	}

	// Replace all disallowed file path characters (both Windows and Unix) so we can safely use URL as directory name
	disallowedChars := []string{"/", "\\", ">", "<", ":", "|", "?", "*"}
	result := u.Host + u.EscapedPath()
	for _, c := range disallowedChars {
		result = strings.Replace(result, c, "-", -1)
	}
	return result, nil
}
*/

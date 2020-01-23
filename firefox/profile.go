package firefox

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// GetProfiles finds all the Firefox profiles on the currently logged into user account.
// The path currently works only on Macos.
func GetProfiles() ([]*Profile, error) {
	profilesPath := filepath.Join(
		os.Getenv("HOME"), "Library", "ApplicationSupport", "Firefox", "Profiles")

	dirs, err := ioutil.ReadDir(profilesPath)
	if err != nil {
		return nil, err
	}

	var profiles []*Profile
	for _, dir := range dirs {
		p := filepath.Join(profilesPath, dir.Name(), "places.sqlite")
		if _, err := os.Stat(p); err == nil {
			profiles = append(profiles, &Profile{Name: dir.Name(), dbFile: p})
		}
	}

	if len(profiles) < 1 {
		return nil, fmt.Errorf("no profiles found under %s", profilesPath)
	}
	return profiles, nil
}

// Profile represents a Firefox profile.
type Profile struct {
	Name   string
	dbFile string
}

// getTmpFile takes the name of an existing file, copies it to the system's tmp
// directory and returns the name of the newly created file.
func getTmpFile(inFile string) (string, error) {
	tmpFileName := filepath.Join(os.TempDir(), "ffprofile.sqlite")

	srcFile, err := os.Open(inFile)
	if err != nil {
		return "", fmt.Errorf("opening source file: %w", err)
	}
	defer srcFile.Close()

	tmpFile, err := os.Create(tmpFileName)
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	defer tmpFile.Close()
	_, err = io.Copy(tmpFile, srcFile)
	if err != nil {
		return "", fmt.Errorf("copying file: %w", err)
	}
	return tmpFileName, nil
}

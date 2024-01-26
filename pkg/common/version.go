package common

// NAME of the App
var NAME = "aws-nuke"

// SUMMARY of the Version
var SUMMARY = "3.0.0-beta.4"

// BRANCH of the Version
var BRANCH = "dev"

var COMMIT = "dirty"

// AppVersion --
var AppVersion AppVersionInfo

// AppVersionInfo --
type AppVersionInfo struct {
	Name    string
	Branch  string
	Summary string
	Commit  string
}

func init() {
	AppVersion = AppVersionInfo{
		Name:    NAME,
		Branch:  BRANCH,
		Summary: SUMMARY,
		Commit:  COMMIT,
	}
}

package common

import "fmt"

// NAME of the App
var NAME = "aws-nuke"

// SUMMARY of the Version
var SUMMARY = "3.0.0-dev"

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

func (a *AppVersionInfo) String() string {
	return fmt.Sprintf("%s - %s - %s", a.Name, a.Summary, a.Commit)
}

func init() {
	AppVersion = AppVersionInfo{
		Name:    NAME,
		Branch:  BRANCH,
		Summary: SUMMARY,
		Commit:  COMMIT,
	}
}

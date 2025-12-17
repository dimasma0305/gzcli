// Package other provides specialized template generators for various CTF-related files
package other

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"

	"github.com/dimasma0305/gzcli/internal/template"
)

// ReadFlag generates a read flag template at the destination
func ReadFlag(destination string) []error {
	return template.TemplateFSToDestination("templates/others/readflag", "", destination)
}

// Writeup generates a writeup template at the destination
func Writeup(destination string, info any) []error {
	return template.TemplateFSToDestination("templates/others/writeup", info, destination)
}

// POC generates a proof-of-concept template at the destination
func POC(destination string, info any) []error {
	return template.TemplateFSToDestination("templates/others/poc", info, destination)
}

// JavaExploitationPlus generates a Java exploitation template at the destination
func JavaExploitationPlus(destination string, info any) []error {
	return template.TemplateFSToDestination("templates/others/java-exploit-plus", info, destination)
}

// CTFInfo contains configuration information for CTF template generation
type CTFInfo struct {
	XorKey         string
	PublicEntry    string
	DiscordWebhook string
	URL            string
	Username       string
	Password       string
	Workspace      string
	RootFolder     string
}

// EventInfo contains configuration information for event template generation
type EventInfo struct {
	Title string
	Start string
	End   string
}

func randomize(n int) string {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

// CTFTemplate generates a complete CTF template structure at the destination
func CTFTemplate(destination string, info any) []error {
	var url, publicEntry, discordWebhook, workspace string

	// Extract values from info map
	if infoMap, ok := info.(map[string]string); ok {
		url = infoMap["url"]
		publicEntry = infoMap["publicEntry"]
		discordWebhook = infoMap["discordWebhook"]
		workspace = infoMap["workspace"]
	}

	// Generate server configuration (.gzctf/)
	absDest, err := filepath.Abs(destination)
	if err != nil {
		absDest = destination
	}

	ctfInfo := &CTFInfo{
		XorKey:         randomize(16),
		Username:       "admin",
		Password:       "ADMIN" + randomize(16) + "ADMIN",
		URL:            url,
		PublicEntry:    publicEntry,
		DiscordWebhook: discordWebhook,
		Workspace:      workspace,
		RootFolder:     absDest,
	}

	// Generate .gzctf/ directory with server files
	errs := template.TemplateFSToDestination("templates/others/ctf-template", ctfInfo, destination)

	// Note: Events are not created automatically.
	// Users should run 'gzcli event create' to create their first event.

	return errs
}

// EventTemplate generates an event directory structure with .gzevent file
func EventTemplate(destination, eventName string, info any) []error {
	var title, start, end string

	// Extract values from info map
	if infoMap, ok := info.(map[string]string); ok {
		title = infoMap["title"]
		start = infoMap["start"]
		end = infoMap["end"]
	}

	eventInfo := &EventInfo{
		Title: title,
		Start: start,
		End:   end,
	}

	eventDest := fmt.Sprintf("%s/events/%s", destination, eventName)

	// Generate event structure from template (includes .gzevent, categories, .example/, .structure/)
	errs := template.TemplateFSToDestination("templates/others/event-template", eventInfo, eventDest)

	return errs
}

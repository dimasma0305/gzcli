// Package other provides specialized template generators for various CTF-related files
package other

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

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

func getUserInput(str string) string {
	var input string
	fmt.Print(str)
	_, _ = fmt.Scanln(&input)
	return input
}

// CTFTemplate generates a complete CTF template structure at the destination
func CTFTemplate(destination string, info any) []error {
	var url, publicEntry, discordWebhook, eventName string

	// Try to extract values from info map if provided
	if infoMap, ok := info.(map[string]string); ok {
		url = infoMap["url"]
		publicEntry = infoMap["publicEntry"]
		discordWebhook = infoMap["discordWebhook"]
		eventName = infoMap["eventName"]
	}

	// Fall back to user input if values are not provided
	if url == "" {
		url = getUserInput("URL: ")
	}
	if publicEntry == "" {
		publicEntry = getUserInput("Public Entry: ")
	}
	if discordWebhook == "" {
		discordWebhook = getUserInput("Discord Webhook (optional): ")
	}
	if eventName == "" {
		eventName = getUserInput("Event Name (e.g., ctf2024): ")
		if eventName == "" {
			eventName = "default-event"
		}
	}

	// Generate server configuration (.gzctf/)
	ctfInfo := &CTFInfo{
		XorKey:         randomize(16),
		Username:       "admin",
		Password:       "ADMIN" + randomize(16) + "ADMIN",
		URL:            url,
		PublicEntry:    publicEntry,
		DiscordWebhook: discordWebhook,
	}

	// Generate .gzctf/ directory with server files
	errs := template.TemplateFSToDestination("templates/others/ctf-template", ctfInfo, destination)
	if len(errs) > 0 {
		return errs
	}

	// Generate event structure (events/[name]/)
	eventInfo := &EventInfo{
		Title: "Example CTF 2024",
		Start: "2024-10-11T12:00:00+00:00",
		End:   "2024-10-13T12:00:00+00:00",
	}

	eventDest := fmt.Sprintf("%s/events/%s", destination, eventName)
	eventErrs := template.TemplateFSToDestination("templates/others/event-template", eventInfo, eventDest)
	if len(eventErrs) > 0 {
		return append(errs, eventErrs...)
	}

	// Set this event as the default
	defaultEventFile := fmt.Sprintf("%s/.gzcli/current-event", destination)
	if err := template.WriteFile(defaultEventFile, []byte(eventName)); err != nil {
		return append(errs, err)
	}

	return errs
}

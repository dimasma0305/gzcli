package uploadserver

import (
	"fmt"
	"io/fs"

	templatepkg "github.com/dimasma0305/gzcli/internal/template"
)

const (
	templateBasePath = "templates/others/event-template/.example"

	templateStaticContainerPath                     = templateBasePath + "/static-container"
	templateStaticAttachmentPath                    = templateBasePath + "/static-attachment"
	templateStaticAttachmentWithComposeLauncherPath = templateBasePath + "/static-attachment-with-compose-launcher"
)

var templateFS fs.FS = templatepkg.File

type challengeTemplate struct {
	Slug       string
	Name       string
	SourcePath string
	Summary    string
}

type templateInfo struct {
	Slug    string
	Name    string
	Summary string
}

var challengeTemplates = []challengeTemplate{
	{
		Slug:       "static-container",
		Name:       "Static Container",
		SourcePath: templateStaticContainerPath,
		Summary:    "Container-based deployment with Dockerfile (per-team container).",
	},
	{
		Slug:       "static-attachment-with-compose-launcher",
		Name:       "Static Attachment (Compose Launcher)",
		SourcePath: templateStaticAttachmentWithComposeLauncherPath,
		Summary:    "Attachment challenge packaged with docker-compose launcher scripts (shared container).",
	},
	{
		Slug:       "static-attachment",
		Name:       "Static Attachment",
		SourcePath: templateStaticAttachmentPath,
		Summary:    "Minimal attachment-only challenge.",
	},
}

func getTemplateBySlug(slug string) (challengeTemplate, bool) {
	for _, tpl := range challengeTemplates {
		if tpl.Slug == slug {
			return tpl, true
		}
	}
	return challengeTemplate{}, false
}

func listTemplateInfo() []templateInfo {
	infos := make([]templateInfo, 0, len(challengeTemplates))
	for _, tpl := range challengeTemplates {
		infos = append(infos, templateInfo{
			Slug:    tpl.Slug,
			Name:    tpl.Name,
			Summary: tpl.Summary,
		})
	}
	return infos
}

func ensureTemplatePaths() error {
	for _, tpl := range challengeTemplates {
		entries, err := fs.ReadDir(templateFS, tpl.SourcePath)
		if err != nil || len(entries) == 0 {
			return fmt.Errorf("template %s is missing or empty", tpl.Slug)
		}
	}
	return nil
}

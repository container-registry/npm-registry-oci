package models

import "encoding/json"

type Package struct {
	ID          string                `json:"_id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	DistTags    map[string]string     `json:"dist-tags"`
	Versions    map[string]Version    `json:"versions"`
	Access      json.RawMessage       `json:"access"`
	Attachments map[string]Attachment `json:"_attachments"`
}

type Version struct {
	Name           string            `json:"name"`
	Version        string            `json:"version"`
	Description    string            `json:"description"`
	Main           string            `json:"main"`
	Scripts        map[string]string `json:"scripts"`
	Author         string            `json:"author"`
	License        string            `json:"license"`
	ID             string            `json:"_id"`
	Readme         string            `json:"readme"`
	ReadmeFileName string            `json:"readmeFilename"`
	GitHead        string            `json:"gitHead"`
	NodeVersion    string            `json:"_nodeVersion"`
	NPMVersion     string            `json:"_npmVersion"`
	Dist           Dist              `json:"dist"`
}

type Dist struct {
	Integrity string `json:"integrity"`
	SHASum    string `json:"shasum"`
	Tarball   string `json:"tarball"`
}

type Attachment struct {
	ContentType string `json:"content_type"`
	Data        string `json:"data"`
	Length      int    `json:"length"`
}

/*
Package minecraft wraps the minecraft versioning api
*/
package minecraft

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type DownloadMetadata struct {
	SHA1 string `json:"sha1"`
	Size int    `json:"sha"`
	URL  string `json:"url"`
}

type VersionMetadata struct {
	Downloads map[string]*DownloadMetadata
}

type Version struct {
	Id          string `json:"id"`
	Type        string `json:"type"`
	Time        string `json:"time"`
	ReleaseTime string `json:"releaseTime"`
	URL         string `json:"url"`
}

type VersionManifest struct {
	Latest struct {
		Release  string `json:"release"`
		Snapshot string `json:"snapshot"`
	}
	Versions []Version
}

func (v *VersionManifest) GetLatestRelease() *Version {
	for _, version := range v.Versions {
		if version.Id == v.Latest.Release {
			return &version
		}
	}
	return nil
}

func (v *VersionManifest) GetRelease(id string) *Version {
	for _, version := range v.Versions {
		if version.Id == id {
			return &version
		}
	}
	return nil
}

const VERSION_MANIFEST_URL = "https://launchermeta.mojang.com/mc/game/version_manifest.json"

var client = &http.Client{Timeout: 10 * time.Second}

// GetVersionManifest gets the latest version manifest
func GetVersionManifest() (*VersionManifest, error) {
	r, err := client.Get(VERSION_MANIFEST_URL)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var manifest VersionManifest
	err = json.NewDecoder(r.Body).Decode(&manifest)
	if err != nil {
		return nil, err
	}

	return &manifest, nil
}

// GetMetadata returns the VersionMetadata for this version
func (v *Version) GetMetadata() (*VersionMetadata, error) {
	r, err := client.Get(v.URL)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var meta VersionMetadata
	err = json.NewDecoder(r.Body).Decode(&meta)
	if err != nil {
		return nil, err
	}

	return &meta, nil
}

// Get downloads the file defined by this DownloadMetadata
func (d *DownloadMetadata) Get(dst io.Writer) error {
	resp, err := client.Get(d.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(dst, resp.Body)
	return err
}

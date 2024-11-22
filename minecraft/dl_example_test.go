package minecraft_test

import (
	"fmt"

	"github.com/b1naryth1ef/sheath/minecraft"
)

func ExampleGetVersionManifest() {
	manifest, err := minecraft.GetVersionManifest()
	if err != nil {
		panic(err)
	}

	version := manifest.GetRelease("1.21.3")
	if version == nil {
		panic("version 1.21.3 not found")
	}

	fmt.Printf("%s %s\n", version.ReleaseTime, version.URL)
	// Output: 2024-10-23T12:28:15+00:00 https://piston-meta.mojang.com/v1/packages/f36ca88c20550b23ce560b53a20a5456cfd10be8/1.21.3.json
}

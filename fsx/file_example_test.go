package fsx_test

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/b1naryth1ef/sheath/fsx"
)

func ExampleNewFileInfo() {
	var buffer bytes.Buffer

	exampleFileData := "hello :3"
	exampleFileInfo := fsx.NewFileInfo("test.txt", len(exampleFileData))

	tw := tar.NewWriter(&buffer)
	header, _ := tar.FileInfoHeader(exampleFileInfo, "")
	tw.WriteHeader(header)
	io.Copy(tw, strings.NewReader(exampleFileData))

	fmt.Printf("Tar Size: %d\n", buffer.Len())
	// Output: Tar Size: 520
}

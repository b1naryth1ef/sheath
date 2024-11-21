package fsx_test

import (
	"archive/tar"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/b1naryth1ef/sheath/fsx"
	"github.com/stretchr/testify/assert"
)

func TestFile(t *testing.T) {
	var buffer bytes.Buffer

	exampleData := "hello :3"
	tw := tar.NewWriter(&buffer)

	header, err := tar.FileInfoHeader(fsx.NewFileInfo("test.txt", len(exampleData)), "")
	assert.Nil(t, err)

	err = tw.WriteHeader(header)
	assert.Nil(t, err)

	_, err = io.Copy(tw, strings.NewReader(exampleData))
	assert.Nil(t, err)
	assert.Equal(t, buffer.Len(), 520)
}

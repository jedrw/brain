package editor

import (
	"os"
	"os/exec"
	"path"
)

func New(filePath string, initialContent []byte) ([]byte, string, error) {
	newFileName := path.Base(filePath)
	tempFile, err := os.CreateTemp("", newFileName)
	if err != nil {
		return []byte{}, tempFile.Name(), err
	}

	_, err = tempFile.Write(initialContent)
	if err != nil {
		return []byte{}, tempFile.Name(), err
	}

	editor, isSet := os.LookupEnv("EDITOR")
	if !isSet {
		editor = "vi"
	}

	editorCmd := exec.Command(editor, tempFile.Name())
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr
	editorCmd.Stdin = os.Stdin
	err = editorCmd.Run()
	if err != nil {
		os.Remove(tempFile.Name())
		return []byte{}, tempFile.Name(), err
	}

	bytes, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return bytes, tempFile.Name(), err
	}

	return bytes, tempFile.Name(), nil
}

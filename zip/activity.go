// Package zip implements activities for reading and writing of zip format files
package zip

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/logger"
)

const (
	ivAction          = "action"
	//ivRemoveFile      = "removeFile"
	ivSourceFile      = "sourceFile"
	ivTargetDirectory = "targetDirectory"
	ovResult          = "result"
)

// log is the default package logger
var log = logger.GetLogger("activity-zip")

// MyActivity is a stub for your Activity implementation
type MyActivity struct {
	metadata *activity.Metadata
}

// NewActivity creates a new activity
func NewActivity(metadata *activity.Metadata) activity.Activity {
	return &MyActivity{metadata: metadata}
}

// Metadata implements activity.Activity.Metadata
func (a *MyActivity) Metadata() *activity.Metadata {
	return a.metadata
}

// Eval implements activity.Activity.Eval
func (a *MyActivity) Eval(context activity.Context) (done bool, err error) {

	// Get the action
	action := context.GetInput(ivAction).(string)
	sourceFile := context.GetInput(ivSourceFile).(string)
	targetDirectory := context.GetInput(ivTargetDirectory).(string)

	// See which action needs to be taken
	switch action {
	case "unzip":
		// Extract file to disk
		err := Unzip(sourceFile, targetDirectory)
		if err != nil {
			// Set the output value in the context
			context.SetOutput(ovResult, err.Error())
			return true, err
		}

		// Set the output value in the context
		context.SetOutput(ovResult, "OK")
		return true, nil
	}

	// Set the output value in the context
	context.SetOutput(ovResult, "NOK")

	return true, nil
}

// Unzip will decompress a zip archive, moving all files and folders
func Unzip(src string, dest string) (error) {

    var filenames []string

    r, err := zip.OpenReader(src)
    if err != nil {
        return err
    }
    defer r.Close()

    for _, f := range r.File {

        // Store filename/path for returning and using later on
        fpath := filepath.Join(dest, f.Name)

        // Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
        if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
            return fmt.Errorf("%s: illegal file path", fpath)
        }

        filenames = append(filenames, fpath)

        if f.FileInfo().IsDir() {
            // Make Folder
            os.MkdirAll(fpath, os.ModePerm)
            continue
        }

        // Make File
        if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
            return err
        }

        outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
        if err != nil {
            return err
        }

        rc, err := f.Open()
        if err != nil {
            return err
        }

        _, err = io.Copy(outFile, rc)

        // Close the file without defer to close before next iteration of loop
        outFile.Close()
        rc.Close()

        if err != nil {
            return err
        }
    }
    return nil
}
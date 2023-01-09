package cmdexec

import (
	"bytes"
	"os/exec"

	"github.com/sirupsen/logrus"
)

func Exec(command string, args ...string) (error, string, string) {
	cmd := exec.Command(command, args...)
	var out bytes.Buffer
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	cmd.Stdout = &out
	logrus.Tracef("Executing: %s\n", cmd.String())
	err := cmd.Run()
	if err == nil {
		logrus.Tracef("Execute successful")
	} else {
		logrus.Tracef("Execute failed %s", err.Error())
	}
	return err, out.String(), stdErr.String()
}

func ExecInFolder(folder, command string, args ...string) (error, string, string) {
	cmd := exec.Command(command, args...)
	cmd.Dir = folder
	var out bytes.Buffer
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	cmd.Stdout = &out
	logrus.Tracef("Executing: %s\n", cmd.String())
	err := cmd.Run()
	if err == nil {
		logrus.Tracef("Execute successful")
	} else {
		logrus.Tracef("Execute failed %s", err.Error())
	}
	return err, out.String(), stdErr.String()
}

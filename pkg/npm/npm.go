package npm

import (
	"os/exec"
)

func Publish() error {
	return exec.Command(faasCliExe, "publish")
}

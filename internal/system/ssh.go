package system

import (
	"os"
	"path"
	"strings"
)

const (
	sshHomeDir             = ".ssh"
	sshHomeDirPerm         = 0700
	authorizedKeysFileName = "authorized_keys"
	authorizedKeysFilePerm = 0600
)

// SetAuthorizedKeys overwrites the file that stores authorized SSH keys for the user with the specified keys.
func SetAuthorizedKeys(username string, keys []string) error {
	user, err := GetUser(username)
	if err != nil {
		return err
	}
	sshDir := path.Join(user.HomeDir, sshHomeDir)
	if _, err := os.Stat(sshDir); err != nil {
		if os.IsNotExist(err) {
			if err = os.Mkdir(sshDir, sshHomeDirPerm); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	if err := os.Chmod(sshDir, sshHomeDirPerm); err != nil {
		return nil
	}
	if err := os.Chown(sshDir, user.Uid, user.Gid); err != nil {
		return err
	}
	keysPath := path.Join(sshDir, authorizedKeysFileName)
	content := strings.Join(keys, "\n") + "\n"
	if err := os.WriteFile(keysPath, []byte(content), authorizedKeysFilePerm); err != nil {
		return err
	}
	return os.Chown(keysPath, user.Uid, user.Gid)
}

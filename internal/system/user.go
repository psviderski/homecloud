package system

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// UserEntry represents a user entry in /etc/passwd file.
type UserEntry struct {
	Username string
	Uid      int
	Gid      int
	HomeDir  string
}

// GetUser parses /etc/passwd file and returns information about the system user.
func GetUser(username string) (UserEntry, error) {
	passwd, err := os.Open("/etc/passwd")
	if err != nil {
		return UserEntry{}, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer passwd.Close()
	scanner := bufio.NewScanner(passwd)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Split(line, ":")
		if len(fields) != 7 || fields[0] != username {
			// Skip a potentially corrupted entry or the wrong user.
			continue
		}
		uid, err := strconv.Atoi(fields[2])
		if err != nil {
			return UserEntry{}, fmt.Errorf("invalid UID for %s user in /etc/passwd", username)
		}
		gid, err := strconv.Atoi(fields[3])
		if err != nil {
			return UserEntry{}, fmt.Errorf("invalid GID for %s user in /etc/passwd", username)
		}
		return UserEntry{
			Username: fields[0],
			Uid:      uid,
			Gid:      gid,
			HomeDir:  fields[5],
		}, nil
	}
	if err := scanner.Err(); err != nil {
		return UserEntry{}, err
	}
	return UserEntry{}, fmt.Errorf("cannot find %s user in /etc/passwd", username)
}

func SetPassword(username, password string) error {
	cmd := exec.Command("chpasswd")
	if strings.HasPrefix(password, "$") {
		cmd.Args = append(cmd.Args, "-e")
	}
	cmd.Stdin = strings.NewReader(fmt.Sprintf("%s:%s", username, password))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set password for user %q: %s: %s", username, err, out)
	}
	return nil
}

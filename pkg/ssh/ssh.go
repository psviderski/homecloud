package ssh

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"strings"
)

// AuthorizedKeyFromPrivate creates an SSH public authorized key corresponding to the private key.
func AuthorizedKeyFromPrivate(key []byte) (string, error) {
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		if _, ok := err.(*ssh.PassphraseMissingError); !ok {
			return "", err
		}
		// TODO: prompt password for the private key.
		return "", fmt.Errorf("SSH private key with passphrase is not supported yet: %w", err)
	}
	return strings.TrimSpace(string(ssh.MarshalAuthorizedKey(signer.PublicKey()))), nil
}

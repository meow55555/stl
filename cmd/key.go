package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"

	"github.com/meow55555/stl/internal/util"
)

const debug = false

func keyCmd(cmd *cobra.Command, args []string) error {
	key, err := util.SSHKey()
	if err != nil {
		return err
	}
	fmt.Printf("%s", key)
	return nil
}

func keyRenewCmd(cmd *cobra.Command, args []string) error {
	configPath, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	configPath, err = util.EnsureDir(path.Join(configPath, "stl"))
	if err != nil {
		return err
	}

	sshKeyPath := path.Join(configPath, "stlKey")

	// remove old ones
	if err := os.Remove(sshKeyPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.Remove(sshKeyPath + ".pub"); err != nil && !os.IsNotExist(err) {
		return err
	}

	// generate new
	if _, err := exec.Command("ssh-keygen", "-f", sshKeyPath).CombinedOutput(); err != nil {
		return err
	}

	return nil
}

const (
	stlControlStart = "# below is controlled by stl, do not change and words"
	stlControlEnd   = "# end of stl controlled zone"
)

func keyAddCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("Miss some args")
	}

	key := args[0]
	if !debug {
		_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key))
		if err != nil {
			return err
		}
	}

	sshPath, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	if debug {
		sshPath = "./"
	}

	sshPath, err = util.EnsureDir(path.Join(sshPath, ".ssh"))
	if err != nil {
		return err
	}

	authKeyPath := path.Join(sshPath, "authorized_keys")

	// read authorized_keys
	authKeysByte, err := ioutil.ReadFile(authKeyPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if strings.Contains(string(authKeysByte), key+"\n") {
		return nil
	}

	authKey := strings.Split(string(authKeysByte), "\n")
	newAuthKey := []string{}

	added := false
	for _, v := range authKey {
		newAuthKey = append(newAuthKey, v)
		if v == stlControlStart {
			added = true
			newAuthKey = append(newAuthKey, key)
		}
	}

	if !added {
		newAuthKey = append(newAuthKey, stlControlStart, key, stlControlEnd)
	}

	ioutil.WriteFile(authKeyPath, []byte(strings.Join(newAuthKey, "\n")), 0600)

	return nil
}

func keyRemoveCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("Miss some args")
	}

	hashed := args[0]

	sshPath, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	if debug {
		sshPath = "./"
	}

	sshPath, err = util.EnsureDir(path.Join(sshPath, ".ssh"))
	if err != nil {
		return err
	}

	authKeyPath := path.Join(sshPath, "authorized_keys")

	// read authorized_keys
	authKeysByte, err := ioutil.ReadFile(authKeyPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	authKey := strings.Split(string(authKeysByte), "\n")
	newAuthKey := []string{}

	for _, v := range authKey {
		if util.Hash(v) != hashed {
			newAuthKey = append(newAuthKey, v)
		}
	}

	ioutil.WriteFile(authKeyPath, []byte(strings.Join(newAuthKey, "\n")), 0600)

	return nil
}

func init() {
	cmd := &cobra.Command{
		Use:   `key`,
		Short: "Return ssh public key",
		RunE:  keyCmd,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   `renew`,
		Short: `Renew your ssh key(this will generate a new ID)(for client)`,
		RunE:  keyRenewCmd,
	}, &cobra.Command{
		Use:   `add <key>`,
		Short: `Add a user's ssh public key(for server)`,
		RunE:  keyAddCmd,
	}, &cobra.Command{
		Use:   `remove <keySHA256>`,
		Short: `Remove key be its sha256 hash`,
		RunE:  keyRemoveCmd,
	})

	rootCmd.AddCommand(cmd)
}

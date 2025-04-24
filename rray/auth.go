package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/charmbracelet/huh"
	"golang.org/x/crypto/ssh"
)

func getAuthMethods(remote string) []ssh.AuthMethod {
	var authMethods []ssh.AuthMethod
	for _, auth := range hostsFile.StoredAuth[remote] {
		switch (auth.Type) {
		case "password":
			authMethods = append(authMethods, ssh.Password(auth.Value))
		case "publickey":
			hmdir, err := os.UserHomeDir()
			if err != nil {
				log.Fatalf("unable to read get home dir")
			}
			key, err := os.ReadFile(path.Join(hmdir, "/.ssh/id_rsa"))
			if err != nil {
				fmt.Println(serror.Render("Failed reading your ssh private key. If you know you have generated one, make sure it's accessible at " + path.Join(hmdir, "/.ssh/id_rsa") + "."))
				os.Exit(1)
			}
			
			var signer ssh.Signer
			if auth.RequiresPassphrase {
				var passphrase string

				if huh.NewInput().
				Title("Please enter your SSH passphrase: ").
				Value(&passphrase).
				Run() != nil {
					log.Fatal("Failed asking for passphrase")
				}
				_signer, err := ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))
				if err != nil {
					fmt.Println(serror.Render("Failed parsing private key using passphrase, is it the correct passphrase?"))
					os.Exit(1)
				}
				signer = _signer
			} else {
				_signer, err := ssh.ParsePrivateKey(key)
				if err != nil {
					log.Fatalf("Failed parsing private key.")
				}
				signer = _signer

			}
			authMethods = append(authMethods, ssh.PublicKeys(signer))
		case "keyboard-interactive": //semi following https://www.rfc-editor.org/rfc/rfc4256.html#section-3.3
			//TODO: filter out control characters from fields according to the spec
			authMethods = append(authMethods, ssh.KeyboardInteractive(
			func(name, instruction string, questions []string, echos []bool) (answers []string, err error) {
				var fields []huh.Field
				values := make([]string, len(questions))

				for indx, question := range questions {
					echoMode := huh.EchoModeNormal
					if (!echos[indx]) {
						echoMode = huh.EchoModePassword
					}

					fields = append(fields,
						huh.NewInput().
						Title(question).
						Value(&values[indx]).
						EchoMode(echoMode))
				}
				if len(fields) == 0 {
					return nil, nil
				}

				if name == "" {
					name = "Authenticating via keyboard interactive"
				}
				if instruction == "" {
					name = "You'll be asked one or more questions to authenticate yourself."
				}

				form := huh.NewForm(
					huh.NewGroup(
						huh.NewNote().
						Title(name).
						Description(instruction).
						Next(true).
						NextLabel("Next"),
					),
					huh.NewGroup(
						fields...
					),
				)
				ferr := form.Run()
				if ferr != nil {
					return nil, ferr
				}
				
				return values, nil
			}))
		}
	}

	return authMethods

}
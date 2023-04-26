package server

import (
          `bytes`
          `errors`
          `fmt`
          `log`
          `os`
          
          `golang.org/x/crypto/ssh`
          
          `SSHChatServer/options`
)

func compareKeyWithWhitelist(key ssh.PublicKey) error {
          entites, err := os.ReadDir(options.Settings.Whitelist)
          if err != nil {
                    return err
          }
          for entity := range entites {
                    filebytes, err := os.ReadFile(fmt.Sprintf("%s/%s", options.Settings.Whitelist, entites[entity].Name()))
                    if err != nil {
                              return err
                    }
                    pub_key, _, _, _, err := ssh.ParseAuthorizedKey(filebytes)
                    if err != nil {
                              continue
                    }
                    if bytes.Equal(key.Marshal(), pub_key.Marshal()) {
                              return nil
                    }
          }
          return errors.New("no entry allowed")
}

func validateKey(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
          err := compareKeyWithWhitelist(key)
          if err != nil {
                    log.Printf("Failed login attempt from %s for user '%s' client: %s\n", conn.RemoteAddr(), conn.User(), conn.ClientVersion())
                    return nil, err
          }
          return nil, nil
}

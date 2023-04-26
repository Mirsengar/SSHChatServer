package server

import (
          `bufio`
          `errors`
          `fmt`
          `os`
          `strings`
          `time`
          
          `golang.org/x/crypto/bcrypt`
          `golang.org/x/term`
)

var passwords_file string = ".shadow"

func passwordRequest(term *term.Terminal, username string) error {
          if userExists(username) {
                    attempts := 3
                    for {
                              pass, err := term.ReadPassword("Enter your password: ")
                              if err != nil {
                                        return err
                              }
                              ok, err := checkPassword(username, pass)
                              if err != nil {
                                        return err
                              }
                              if ok {
                                        return nil
                              }
                              time.Sleep(time.Second * 1)
                              attempts -= 1
                              if attempts <= 0 {
                                        return errors.New("Invalid password")
                              }
                              term.Write([]byte("Permission denied, please try again.\n"))
                    }
          } else {
                    for {
                              pass, err := term.ReadPassword("Enter new password: ")
                              if err != nil {
                                        return err
                              }
                              pass_c, err := term.ReadPassword("Сonfirm your password: ")
                              if err != nil {
                                        return err
                              }
                              if pass != pass_c {
                                        term.Write([]byte("Passwords don't match\n"))
                                        continue
                              }
                              err = addPassword(username, pass)
                              if err != nil {
                                        return err
                              }
                              term.Write([]byte("\nYour password saved\n\n"))
                              return nil
                    }
          }
}

func updatePasswordRequest(term *term.Terminal, username string) error {
          for {
                    pass, err := term.ReadPassword("Enter new password: ")
                    if err != nil {
                              return err
                    }
                    pass_c, err := term.ReadPassword("Сonfirm your password: ")
                    if err != nil {
                              return err
                    }
                    if pass != pass_c {
                              term.Write([]byte("Passwords don't match\n"))
                              continue
                    }
                    if userExists(username) {
                              err = updateUserPassword(username, pass)
                              if err != nil {
                                        return err
                              }
                    } else {
                              err = addPassword(username, pass)
                              if err != nil {
                                        return err
                              }
                    }
                    term.Write([]byte("\nYour password updated\n\n"))
                    break
          }
          return nil
}

func checkPassword(username, password string) (bool, error) {
          file, err := os.OpenFile(passwords_file, os.O_RDWR|os.O_CREATE, 0600)
          if err != nil {
                    return false, err
          }
          defer file.Close()
          file_scanner := bufio.NewScanner(file)
          file_scanner.Split(bufio.ScanLines)
          for file_scanner.Scan() {
                    password_data := strings.Split(file_scanner.Text(), ":")
                    if password_data[0] == username {
                              return CheckPasswordHash(password, password_data[1]), nil
                    }
          }
          return false, errors.New("user not exists")
}

func userExists(username string) bool {
          file, err := os.OpenFile(passwords_file, os.O_RDWR|os.O_CREATE, 0600)
          if err != nil {
                    return false
          }
          defer file.Close()
          file_scanner := bufio.NewScanner(file)
          file_scanner.Split(bufio.ScanLines)
          for file_scanner.Scan() {
                    password_data := strings.Split(file_scanner.Text(), ":")
                    if password_data[0] == username {
                              return true
                    }
          }
          return false
}

func updateUserPassword(username, new_password string) error {
          new_password_hash, err := HashPassword(new_password)
          if err != nil {
                    return err
          }
          file, err := os.OpenFile(passwords_file, os.O_RDWR|os.O_CREATE, 0600)
          if err != nil {
                    return err
          }
          defer file.Close()
          file_scanner := bufio.NewScanner(file)
          file_scanner.Split(bufio.ScanLines)
          var passwords []string
          for file_scanner.Scan() {
                    passwords = append(passwords, file_scanner.Text())
          }
          var configured_string string
          for password_s := range passwords {
                    password_data := strings.Split(passwords[password_s], ":")
                    if password_data[0] == username {
                              password_data[1] = new_password_hash
                    }
                    configured_string = fmt.Sprintf("%s%s:%s\n", configured_string, password_data[0], password_data[1])
          }
          err = file.Truncate(0)
          if err != nil {
                    return err
          }
          _, err = file.Seek(0, 0)
          if err != nil {
                    return err
          }
          _, err = file.WriteString(configured_string)
          if err != nil {
                    return err
          }
          err = file.Sync()
          return err
}
func addPassword(username, password string) error {
          err := checkUsername(username)
          if err != nil {
                    return err
          }
          password_hash, err := HashPassword(password)
          if err != nil {
                    return err
          }
          pass_string := fmt.Sprintf("%s:%s\n", username, password_hash)
          file, err := os.OpenFile(passwords_file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
          defer file.Close()
          file.Write([]byte(pass_string))
          return err
}
func HashPassword(password string) (string, error) {
          bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
          return string(bytes), err
}
func CheckPasswordHash(password, hash string) bool {
          err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
          return err == nil
}
func checkUsername(name string) error {
          if strings.ContainsAny(name, ":\"'`/\\") {
                    return errors.New("Incorrect username. Not use symbols :\"'`/\\")
          }
          return nil
}

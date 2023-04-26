package server

import (
          `fmt`
          `log`
          `net`
          
          "golang.org/x/crypto/ssh"
          `golang.org/x/crypto/ssh/terminal`
          
          `SSHChatServer/rooms`
          `SSHChatServer/utils`
)

func NewServer(privateKey []byte, passwordMode bool, client_auth bool) (*Server, error) {
          signer, err := ssh.ParsePrivateKey(privateKey)
          if err != nil {
                    return nil, err
          }
          config := ssh.ServerConfig{
                    NoClientAuth:      !client_auth,
                    PublicKeyCallback: validateKey,
          }
          config.AddHostKey(signer)
          server := Server{
                    sshConfig:    &config,
                    sshSigner:    &signer,
                    passwordMode: passwordMode,
          }
          return &server, nil
}

func (s *Server) Start(laddr string) (<-chan struct{}, error) {
          socket, err := net.Listen("tcp", laddr)
          if err != nil {
                    return nil, err
          }
          s.socket = &socket
          log.Printf("Listening on %s", laddr)
          rooms.CreateRoom("main")
          go func() {
                    for {
                              conn, err := socket.Accept()
                              if err != nil {
                                        log.Printf("Failed to accept connection, aborting loop: %v", err)
                                        return
                              }
                              sshConn, channels, requests, err := ssh.NewServerConn(conn, s.sshConfig)
                              if err != nil {
                                        log.Printf("Failed to handshake: %v", err)
                                        continue
                              }
                              log.Printf("Connection from: %s, %s, %s", sshConn.RemoteAddr(), sshConn.User(), sshConn.ClientVersion())
                              go ssh.DiscardRequests(requests)
                              go s.handleChannels(channels, sshConn)
                    }
          }()
          return s.done, nil
}

func (s *Server) handleChannels(channels <-chan ssh.NewChannel, conn *ssh.ServerConn) {
          for ch := range channels {
                    if t := ch.ChannelType(); t != "session" {
                              ch.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
                              continue
                    }
                    channel, requests, err := ch.Accept()
                    if err != nil {
                              continue
                    }
                    err = checkUsername(conn.User())
                    if err != nil {
                              channel.Write([]byte(fmt.Sprintf("incorrect username: %v\t", err)))
                              conn.Close()
                    }
                    go func(in <-chan *ssh.Request) {
                              defer channel.Close()
                              for req := range in {
                                        ok := false
                                        switch req.Type {
                                        case "shell":
                                                  if len(req.Payload) == 0 {
                                                            ok = true
                                                  }
                                        case "pty-req":
                                                  ok = true
                                        case "window-change":
                                                  continue // no response
                                        }
                                        req.Reply(ok, nil)
                              }
                    }(requests)
                    go s.handleShell(channel, conn.User())
          }
}

func (s *Server) handleShell(channel ssh.Channel, username string) {
          defer func() {
                    channel.Close()
                    log.Println(username, "disconnected")
          }()
          term := terminal.NewTerminal(channel, fmt.Sprintf("%s > ", username))
          if s.passwordMode || userExists(username) {
                    err := passwordRequest(term, username)
                    if err != nil {
                              return
                    }
          }
          utils.PrintRandomLogo(term)
          var currentRoom *rooms.Room
          currentRoom, err := rooms.JoinRoom("main", rooms.User{Nickname: username, Terminal: term})
          if err != nil {
                    log.Println("Room main not found")
                    term.Write([]byte("Room main not found\n"))
                    return
          }
          for {
                    line, err := term.ReadLine()
                    if err != nil {
                              break
                    }
                    if len(line) > 0 {
                              if string(line[0]) == "/" {
                                        switch line {
                                        case "/exit":
                                                  return
                                        case "/help":
                                                  term.Write([]byte(helpMessage))
                                        case "/new_password":
                                                  err = updatePasswordRequest(term, username)
                                                  if err != nil {
                                                            term.Write([]byte("your password not updated\n"))
                                                  }
                                        default:
                                                  term.Write([]byte(helpMessage))
                                        }
                                        continue
                              }
                              currentRoom.Message <- rooms.Message{FromUser: username, Message: line}
                    }
          }
}

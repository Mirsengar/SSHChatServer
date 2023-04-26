package rooms

import (
          `errors`
          `fmt`
)

var roomPool []*Room = make([]*Room, 0, 5)

func JoinRoom(roomName string, u User) (*Room, error) {
          for room := range roomPool {
                    if roomPool[room].Title == roomName {
                              roomPool[room].Users = append(roomPool[room].Users, u)
                              return roomPool[room], nil
                    }
          }
          return nil, errors.New("room not exists")
}
func CreateRoom(roomName string) {
          newRoom := Room{
                    Message: make(chan Message),
                    Title:   roomName,
          }
          roomPool = append(roomPool, &newRoom)
}
func RoomMessageRouter(r *Room) {
          for message := range r.Message {
                    var msg []byte = []byte(fmt.Sprintf("%s: %s\n", message.FromUser, message.Message))
                    for RoomUser := range r.Users {
                              if r.Users[RoomUser].Nickname != message.FromUser {
                                        r.Users[RoomUser].Terminal.Write(msg)
                              }
                    }
          }
}

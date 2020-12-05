package main

import (
    "context"
    "errors"
    "fmt"
    "io"
    "net"

    "google.golang.org/grpc"
    "log"

    pb "github.com/tk-aria/model/room.pb"
)

// gRPC struct
type server struct {
    rooms []room
}

// チャットルーム
type room struct {
    id string
    contents []message
}

// チャットメッセージ
type message struct {
    author string
    content string
}

// Greet
func (s *server) GreetServer(ctx context.Context, p *pb.GreetRequest) (*pb.GreetMessage, error) {
    log.Printf("Request from: %s", p.Name)
    return &pb.GreetMessage{Msg: fmt.Sprintf("Hello, %s. ", p.Name)}, nil
}

// チャットルームの追加
func (s *server) AddRoom(ctx context.Context, p *pb.RoomRequest) (*pb.RoomInfo, error) {
    log.Printf("Add Room Request")
    // チャットルームをスライスに追加
    s.rooms = append(s.rooms, room{id: p.Id, contents: []message{}})
    // チャットルームの探索
    index, err := searchRooms(s.rooms, p.Id)
    if err != nil {
        return nil, err
    }
    room := s.rooms[index]
    return &pb.RoomInfo{Id: room.id, MessageCount: int32(len(room.contents))}, nil
}

// チャットルームの情報取得
func (s *server) GetRoomInfo(ctx context.Context, p *pb.RoomRequest) (*pb.RoomInfo, error) {
    log.Printf("Get Room Request")
    // チャットルームの探索
    index, err := searchRooms(s.rooms, p.Id)
    if err != nil {
        return nil, err
    }
    room := s.rooms[index]
    return &pb.RoomInfo{Id: room.id, MessageCount: int32(len(room.contents))}, nil
}

// チャットルームの一覧を取得
func (s *server) GetRooms(ctx context.Context, p *pb.Null) (*pb.RoomList, error) {
    log.Printf("Get Rooms Request")
    return &pb.RoomList{Rooms: buildRoomInfo(s.rooms)}, nil
}

// チャットルームへstreamを使いメッセージを送信する
func (s *server) SendMessage(stream pb.HelloGrpc_SendMessageServer) error{
    // 無限ループ
    for {
        // クライアントからメッセージ受信
        m, err := stream.Recv()
        log.Printf("Receive message>> [%s] %s", m.Name, m.Content)
        // EOF、エラーなら終了
        if err == io.EOF {
            // EOFなら接続終了処理
            return stream.SendAndClose(&pb.SendResult{
                Result: true,
            })
        }
        if err != nil {
            return err
        }
        // 終了コマンド
        if m.Content == "/exit" {
            return stream.SendAndClose(&pb.SendResult{
                Result: true,
            })
        }
        // チャットルームの探索
        index, err := searchRooms(s.rooms, m.Id)
        if err != nil {
            return err
        }
        // メッセージの追加
        s.rooms[index].contents = append(
            s.rooms[index].contents,
            message{
                author: m.Name,
                content: m.Content,
            },
        )
    }
}

// チャットルームの新着メッセージをstreamを使い配信する
func (s *server) GetMessages(p *pb.MessagesRequest, stream pb.HelloGrpc_GetMessagesServer) error{
    // チャットルームの探索
    index, err := searchRooms(s.rooms, p.Id)
    if err != nil {
        return err
    }
    // 対象チャットルーム
    targetRoom := s.rooms[index]
    // 差を使って新着メッセージを検知する
    previousCount := len(targetRoom.contents)
    currentCount := 0
    // 無限ループ
    for {
        targetRoom = s.rooms[index]
        currentCount = len(targetRoom.contents)
        // 現在のmessageCountが前回より多ければ新着メッセージあり
        if previousCount < currentCount {
            msg, _ := latestMessage(targetRoom.contents)
            // クライアントへメッセージ送信
            if err := stream.Send(&pb.Message{Id: targetRoom.id, Name: msg.author, Content: msg.content}); err != nil {
                return err
            }
        }
        previousCount = currentCount
    }
}

func latestMessage(messages []message) (message, error) {
    length := len(messages)
    if length == 0 {
        return message{}, errors.New("Not found")
    }
    return messages[length - 1], nil
}

func buildRoomInfo(rooms []room) ([]*pb.RoomInfo) {
    r := make([]*pb.RoomInfo, 0)
    for _, v := range(rooms) {
        r = append(r, &pb.RoomInfo{Id: v.id, MessageCount: int32(len(v.contents))})
    }
    return r
}

func searchRooms(r []room, id string) (int, error) {
    for i, v := range(r) {
        if v.id == id {
            return i, nil
        }
    }
    return -1, errors.New("Not found")
}

func main() {
    // gRPC
    port := 10000
    lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
    if err != nil {
        log.Fatalf("lfailed to listen: %v", err)
    }
    log.Printf("Run server port: %d", port)
    grpcServer := grpc.NewServer()
    pb.RegisterHelloGrpcServer(grpcServer, &server{rooms: []room{}})
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}

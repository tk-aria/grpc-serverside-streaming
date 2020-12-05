package main

import (
    "bufio"
    "context"
    "io"
    "log"
    "os"
    "time"

    "google.golang.org/grpc"
    pb "github.com/tk-aria/model/room.pb"


const (
    address = "localhost:10000"
    defaultName = "hoge"
)

func greetServer(ctx context.Context, c pb.HelloGrpcClient, name string) error {
    r, err := c.GreetServer(ctx, &pb.GreetRequest{Name: name})
    if err != nil {
        return err
    }
    log.Printf("%s", r.Msg)
    return nil
}

// チャットルームを作成するためのサーバのメソッドを呼び出す
func createRoom(ctx context.Context, c pb.HelloGrpcClient, id string) error {
    r, err := c.AddRoom(ctx, &pb.RoomRequest{Id: id})
    if err != nil {
        return err
    }
    log.Printf("Created room. >> %s", r.Id)
    return nil
}

// チャットルームの情報を取得するためのサーバのメソッドを呼び出す
func getRoom(ctx context.Context, c pb.HelloGrpcClient, id string) error {
    r, err := c.GetRoomInfo(ctx, &pb.RoomRequest{Id: id})
    if err != nil {
        return err
    }
    log.Printf("Room information. >> id: %s, messageCount: %d", r.Id, r.MessageCount)
    return nil
}

// チャットルームの一覧を取得するためのサーバのメソッドを呼び出す
func getRooms(ctx context.Context, c pb.HelloGrpcClient, name string) error {
    r, err := c.GetRooms(ctx, &pb.Null{})
    if err != nil {
        return err
    }
    for _, v := range(r.Rooms) {
        log.Printf("Name: %s, MessageCount: %d", v.Id, v.MessageCount)
    }
    return nil
}

// streamを使いサーバへ連続してチャットメッセージを送信する
func sendMessage(c pb.HelloGrpcClient, id string, name string) error {
    // 標準入力を使ってメッセージを入力
    stdin := bufio.NewScanner(os.Stdin)
    // サーバへstreamを渡す
    stream, err := c.SendMessage(context.Background())
    if err != nil {
        return err
    }
    // 無限ループを使ってで連続してメッセージ送信
    for {
        // 入力待ち
        stdin.Scan()
        text := stdin.Text()
        // サーバへSendRequest型のメッセージを送信
        if err := stream.Send(&pb.SendRequest{Id: id, Name: name, Content: text}); err != nil {
            log.Fatalf("Send failed: %v", err)
        }
        // /exitを入力すると終了
        if text == "/exit" {
            break
        }
    }
    // 接続終了処理
    _, err = stream.CloseAndRecv()
    if err != nil {
        return err
    }
    return nil
}

// streamを使いサーバから連続してメッセージを受信する
func getMessage(c pb.HelloGrpcClient, id string) error {
    req := &pb.MessagesRequest{Id: id}
    // サーバからstreamを受け取る
    stream, err := c.GetMessages(context.Background(), req)
    if err != nil {
        return err
    }
    // 無限ループ
    for {
        // サーバからのメッセージを受信
        msg, err := stream.Recv()
        if err == io.EOF {
            break
        }
        log.Printf("[%s] %s", msg.Name, msg.Content)
        if err != nil {
            return err
        }
    }
    return nil
}

func main() {
    conn, err := grpc.Dial(address, grpc.WithInsecure())
    if err != nil {
        log.Fatalf("did not connect: %v", err)
    }
    defer conn.Close()
    c := pb.NewHelloGrpcClient(conn)

    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()
    // case文で分岐
    if len(os.Args) > 2 {
        switch os.Args[1] {
        case "greet":
            err := greetServer(ctx, c, os.Args[2])
            if err != nil {
                log.Fatalf("Couldn't execute: %v", err)
            }
        case "add":
            err := createRoom(ctx, c, os.Args[2])
            if err != nil {
                log.Fatalf("Couldn't execute: %v", err)
            }
        case "get":
            err := getRoom(ctx, c, os.Args[2])
            if err != nil {
                log.Fatalf("Couldn't execute: %v", err)
            }
        case "list":
            err := getRooms(ctx, c, os.Args[2])
            if err != nil {
                log.Fatalf("Couldn't execute: %v", err)
            }
        case "send":
            err := sendMessage(c, os.Args[2], os.Args[3])
            if err != nil {
                log.Fatalf("Couldn't execute: %v", err)
            }
        case "stream":
            err := getMessage(c, os.Args[2])
            if err != nil {
                log.Fatalf("Couldn't execute: %v", err)
            }
        case "chat":
            // goroutine(非同期処理)を使ってメッセージ受信・表示
            go getMessage(c, os.Args[2])
            // メッセージ送信はmainで実行
            err := sendMessage(c, os.Args[2], os.Args[3])
            if err != nil {
                log.Fatalf("Couldn't execute: %v", err)
            }
        default:
            log.Fatalf("Unknown command.")
        }
    } else {
        log.Fatalf("Need arguments.")
    }
}
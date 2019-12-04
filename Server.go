epackage main

import (
	"net"
	"fmt"
	"os"
	"strings"
	"io"
	"time"
	"io/ioutil"
)

/*
单撩
群撩
上线通知
昵称
*/

type Client struct {
	conn net.Conn
	name string
	addr string
}

var (
	//客户端信息,用昵称为键
	//clientsMap = make(map[string]net.Conn)
	clientsMap = make(map[string]Client)
)

func SHandleError(err error, why string) {
	if err != nil {
		fmt.Println(why, err)
		os.Exit(1)
	}
}

func main() {

	//建立服务端监听
	listener, e := net.Listen("tcp", "127.0.0.1:8888")
	SHandleError(e, "net.Listen")
	defer func() {
		for _, client := range clientsMap {
			client.conn.Write([]byte("all:服务器进入维护状态，大家都洗洗睡吧！"))
		}
		listener.Close()
	}()

	for {
		//循环接入所有女朋友
		conn, e := listener.Accept()
		SHandleError(e, "listener.Accept")
		clientAddr := conn.RemoteAddr()

		//TODO:接收并保存昵称
		buffer := make([]byte, 1024)
		var clientName string
		for {
			n, err := conn.Read(buffer)
			SHandleError(err, "conn.Read(buffer)")
			if n > 0 {
				clientName = string(buffer[:n])
				break
			}
		}
		fmt.Println(clientName + "上线了")

		//TODO:将每一个女朋友丢入map
		client := Client{conn, clientName, clientAddr.String()}
		clientsMap[clientName] = client

		//TODO:给已经在线的用户发送上线通知——使用昵称
		for _, client := range clientsMap {
			client.conn.Write([]byte(clientName + "上线了"))
		}

		//在单独的协程中与每一个具体的女朋友聊天
		go ioWithClient(client)
	}

	//设置优雅退出逻辑

}

//与一个Client做IO
func ioWithClient(client Client) {
	//clientAddr := conn.RemoteAddr().String()
	buffer := make([]byte, 1024)

	for {
		n, err := client.conn.Read(buffer)
		if err != io.EOF {
			SHandleError(err, "conn.Read")
		}

		if n > 0 {
			msg := string(buffer[:n])
			fmt.Printf("%s:%s\n", client.name, msg)

			//将客户端说的每一句话记录在【以他的名字命名的文件里】
			writeMsgToLog(msg, client)

			strs := strings.Split(msg, "#")
			if len(strs) > 1 {
				//all#hello
				//zqd#hello

				//要发送的目标昵称
				targetName := strs[0]
				targetMsg := strs[1]

				//TODO:使用昵称定位目标客户端的Conn
				if targetName == "all" {
					//群发消息
					for _, c := range clientsMap {
						c.conn.Write([]byte(client.name + ":" + targetMsg))
					}
				} else {
					//点对点消息
					for key, c := range clientsMap {
						if key == targetName {
							c.conn.Write([]byte(client.name + ":" + targetMsg))

							//在点对点消息的目标端也记录日志
							go writeMsgToLog(client.name + ":" + targetMsg,c)
							break
						}
					}
				}

			} else {

				//客户端主动下线
				if msg == "exit" {
					//将当前客户端从在线用户中除名
					//向其他用户发送下线通知
					for name, c := range clientsMap {
						if c == client {
							delete(clientsMap, name)
						} else {
							c.conn.Write([]byte(name + "下线了"))
						}
					}
					// strings.Index 在msg语句中，找到log@的第一个下标
				}else if strings.Index(msg,"log@")==0 {
					//log@all
					//log@张全蛋
					filterName := strings.Split(msg, "@")[1]
					//向客户端发送它的聊天日志
					go sendLog2Client(client,filterName)
				} else {
					client.conn.Write([]byte("已阅：" + msg))
				}

			}

		}
	}

}

//向客户端发送它的聊天日志
func sendLog2Client(client Client,filterName string) {
	//读取聊天日志
	logBytes, e := ioutil.ReadFile("D:/BJBlockChain1801/demos/W4/day1/01ChatRoomII/logs/" + client.name + ".log")
	SHandleError(e,"ioutil.ReadFile")

	if filterName != "all"{
		//查找与某个人的聊天记录
		//从内容中筛选出带有【filterName#或filterName:】的行，拼接起来
		logStr := string(logBytes)
		targetStr := ""
		lineSlice := strings.Split(logStr, "\n")
		for _,lineStr := range lineSlice{
			if len(lineStr)>20{
				contentStr := lineStr[20:]
				if strings.Index(contentStr,filterName+"#")==0 || strings.Index(contentStr,filterName+":")==0{
					targetStr += lineStr+"\n"
				}
			}
		}
		client.conn.Write([]byte(targetStr))
	}else{
		//查询所有的聊天记录
		//向客户端发送
		client.conn.Write(logBytes)
	}

}

//将客户端说的一句话记录在【以他的名字命名的文件里】
func writeMsgToLog(msg string, client Client) {
	//打开文件
	file, e := os.OpenFile(
		"D:/BJBlockChain1801/demos/W4/day1/01ChatRoomII/logs/"+client.name+".log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644)
	SHandleError(e, "os.OpenFile")
	defer file.Close()

	//追加这句话
	logMsg := fmt.Sprintln(time.Now().Format("2006-01-02 15:04:05"), msg)
	file.Write([]byte(logMsg))
}

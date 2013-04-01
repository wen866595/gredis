package main

import (
	gredis "./gredis"
	gutil "./util"
	"flag"
	"fmt"
	"net"
	"os"
	"bufio"
	"strings"
	"strconv"
)

func getIpPort() (string) {
    ipFlag := flag.String("ip", "127.0.0.1", "redis server ip .")
    portFlag := flag.String("port", "6379", "redis server port .")
    flag.Parse()
    return *ipFlag + ":" + *portFlag
}

func main() {
    ipport := getIpPort()

	conn, err := net.Dial("tcp", ipport)
	if err != nil {
		println("cann't connect to redis server " + ipport)
		return
	}

    db := 0
	stdin := bufio.NewReader(os.Stdin)
	for {
        if db == 0 {
		    fmt.Printf("redis %s>", ipport)
        } else {
		    fmt.Printf("redis %s[%d]>", ipport, db)
        }

		strLine, err := gutil.ReadLine(stdin)
		if err != nil {
			fmt.Println("read command error, client will quit .")
			break
		}

		strLine = strings.TrimSpace(strLine)
		if strings.EqualFold(strLine, "quit") {
			break
		}

		cmdArg, err := gutil.ParseStrLine(strLine)
		if err != nil {
			fmt.Printf("error command : %v\nPlease input right command .\n", err)
			continue
		}

		cmd, _ := gredis.BuildStringCmd(cmdArg)

		err = gutil.WriteFull(conn, cmd)
		if err != nil {
			println("write cmd error !")
			return
		}

		reply, err := gredis.ReadResponse(conn)
		if err != nil {
			fmt.Printf("error happend : %v", err)
            break
		} else {
            db = checkDb(cmdArg, db)
            println(reply.String())
        }
	}

	conn.Close()
}

func checkDb(cmd []string, db int) int {
    if strings.EqualFold("select", cmd[0]) {
        i, _ := strconv.Atoi(cmd[1])
        return i
    }
    return db
}



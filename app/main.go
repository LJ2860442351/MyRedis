package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"project/MyRedis/util"
	"strings"
	"time"

	"github.com/peterh/liner"
)

var file = "C:\\tmp\\MyRedis\\set.txt"

func main001() {
	fmt.Println("MyRedis is starting....")
	defaultExpiration, _ := time.ParseDuration("0.5h")
	gcInterval, _ := time.ParseDuration("3s")
	c := util.NewCache(defaultExpiration, gcInterval)

	k1 := "hello world"
	expiration, _ := time.ParseDuration("5s")

	//设置一个k1值
	c.Set("k1", k1, expiration)
	//睡眠10s，设置对比，看下数据有没有过期，因为我们上面设置的是5s
	s, _ := time.ParseDuration("10s")
	v, found := c.Get("k1")
	if found {
		fmt.Println("Found k1:", v)
	} else {
		fmt.Println("Not Found k1")
	}
	time.Sleep(s)
	if v, found := c.Get("k1"); found {
		fmt.Println("Found k1:", v)
	} else {
		fmt.Println("Not Found k1")
	}
}

var (
	history_fn = filepath.Join(os.TempDir(), ".liner_example_history")
	names      = []string{"john", "james", "mary", "nancy"}
)

func main() {
	defaultExpiration, _ := time.ParseDuration("0.5h")
	gcInterval, _ := time.ParseDuration("3s")
	c := util.NewCache(defaultExpiration, gcInterval)

	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)

	line.SetCompleter(func(line string) (c []string) {
		for _, n := range names {
			if strings.HasPrefix(n, strings.ToLower(line)) {
				c = append(c, n)
			}
		}
		return
	})

	for {
		name, err := line.Prompt("127.0.0.1:5200>")
		if err == liner.ErrPromptAborted {
			log.Print("Aborted")
			break
		} else if name == "help" || name == "HELP" {
			printHelper()
		} else if name == "quit" || name == "QUIT" {
			fmt.Println("exited!!Thanks for use.")
			break
		} else {
			cmd := strings.Fields(name)
			//set命令
			switch cmd[0] {
			case "set":
				if len(cmd) < 4 {
					c.Set(cmd[1], cmd[2], -1)
				} else {
					expiration, _ := time.ParseDuration(cmd[3])
					c.Set(cmd[1], cmd[2], expiration)
				}
			case "get":
				//get命令
				if v, found := c.Get(cmd[1]); found {
					//如果查到了这个值，直接返回这个值
					fmt.Println(v)
				} else {
					//没有遍历得到的值
					fmt.Println("Not Found:", cmd[1])
				}
			case "all":
				c.All()
			case "load":
				c.LoadFile(file)
			default:
				//其他的命令
				//暂时没有发现的命令
				fmt.Println("Not Found the command", cmd[0])
				printHelper()

			}
			// if cmd[0] == "set" {
			// 	//第4个参数是过期时间
			// 	expiration, _ := time.ParseDuration("999999s")
			// 	c.Set(cmd[1], cmd[2], expiration)
			// } else if cmd[0] == "get" {
			// 	//get命令
			// 	if v, found := c.Get(cmd[1]); found {
			// 		//如果查到了这个值，直接返回这个值
			// 		fmt.Println(v)
			// 	} else {
			// 		//没有遍历得到的值
			// 		fmt.Println("Not Found:", cmd[1])
			// 	}
			// } else {
			// 	//其他的命令
			// 	//暂时没有发现的命令
			// 	fmt.Println("Not Found the command", cmd[0])
			// 	printHelper()
			// }
		}
		if f, err := os.Create(history_fn); err != nil {
			log.Print("Error writing history file: ", err)
		} else {
			line.WriteHistory(f)
			f.Close()
		}
	}
}

func printHelper() {
	help := `Thanks for using MyRedis
	And the command is
	MyRedis-cli
	To get help about command:
		Type: "help <command>" for help on command
	To quit:
		<ctrl+c> or <quit>`
	fmt.Println(help)
}

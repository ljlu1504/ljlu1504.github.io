package main

import (
	"github.com/ljlu1504/ljlu1504.github.io/util"
	"log"
	"fmt"
	"time"
)

var gitCount = 1

func createCmds() []util.Cmd {
	gitCount++
	gifConfig1 := []util.Cmd{
		{"git", []string{"config", "--global", "user.email", "'ljlu1504@gmail.com'"}},
	}
	gifConfig2 := []util.Cmd{
		{"git", []string{"config", "--global", "user.email", "'ljlu1504@gmail.com'"}},
	}
	cmds := []util.Cmd{
		{"git", []string{"config", "--global", "user.name", "'ljlu1504'"}},
		{"git", []string{"stash"}},
		{"git", []string{"pull", "origin", "master"}},
		{"git", []string{"stash", "apply"}},
		{"git", []string{"add", "."}},
		{"git", []string{"status"}},
		{"git", []string{"commit", "-am", "hacknews-update" + time.Now().Format(time.RFC3339)}},
		{"git", []string{"status"}},
		{"git", []string{"push", "origin", "master"}},
		//{"netstat", []string{"-lntp"}},
		//{"free", []string{"-m"}},
		//{"ps", []string{"aux"}},
	}
	if gitCount%2 == 0 {
		cmds = append(gifConfig2, cmds...)
	} else {
		cmds = append(gifConfig1, cmds...)
	}
	return cmds
}

func main() {
	for {
		if err := util.SpiderHackNews(); err != nil {
			log.Println(err)
		}
		if err := util.SpiderHackShows(); err != nil {
			log.Println(err)
		}
		if err := util.ParsemarkdownHacknews(); err != nil {
			log.Println(err)
		}
		fmt.Println("finished spider")
		_, err := util.RunCmds(createCmds())
		if err != nil {
			log.Println(err)
		}
		fmt.Println("finished RunCmds")
		time.Sleep(3 * time.Hour)
	}
}

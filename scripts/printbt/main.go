package main

import (
	"fmt"

	"github.com/xescugc/maze-wars/server/bot"
)

func main() {
	b := bot.Bot{}
	fmt.Println(b.Node().String())
}

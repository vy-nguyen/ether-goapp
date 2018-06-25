package main

import (
	_ "github.com/go-sql-driver/mysql"
	_ "tudo/kstore"
	_ "tudo/models"
	_ "tudo/routers"

	// "github.com/astaxie/beego"
	"tudo/ether"
)

func main() {
	ether.GethMain()
	// beego.Run()
}

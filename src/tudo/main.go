package main

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "tudo/models"
	_ "tudo/routers"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"tudo/ether"
)

func init() {
	fmt.Printf("Init sql database connector")
}

func main() {
	o := orm.NewOrm()
	fmt.Printf("%v", o)

	go ether.GethMain()
	beego.Run()
}

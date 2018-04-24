package models

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"

	_ "github.com/go-sql-driver/mysql"
)

func init() {
	fmt.Printf("Init database...")
	orm.RegisterDriver("mysql", orm.DRMySQL)
	conf := beego.AppConfig
	part := []string{
		conf.String("mysqluser"), ":", conf.String("mysqlpass"),
		"@tcp(", conf.String("mysqlurls"), ":3306)/",
		conf.String("mysqldb"), "?charset=utf8",
	}
	orm.RegisterDataBase("default", "mysql", strings.Join(part, ""))

	orm.RegisterModel(new(Account))
	orm.RegisterModel(new(Transaction))
	orm.RegisterModel(new(AccountKey))

	orm.RunSyncdb("default", false, true)
}

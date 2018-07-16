package models

import (
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"

	_ "github.com/go-sql-driver/mysql"
)

func init() {
	orm.RegisterDriver("mysql", orm.DRMySQL)
	conf := beego.AppConfig
	part := []string{
		conf.String("mysqluser"), ":", conf.String("mysqlpass"),
		"@tcp(", conf.String("mysqlurls"), ":3306)/",
		conf.String("mysqldb"), "?charset=utf8",
	}
	admin := []string{
		conf.String("mysqluser"), ":", conf.String("mysqlpass"),
		"@tcp(", conf.String("mysqlurls"), ":3306)/",
		conf.String("admindb"), "?charset=utf8",
	}
	orm.RegisterDataBase("default", "mysql", strings.Join(part, ""))
	orm.RegisterDataBase("admin", "mysql", strings.Join(admin, ""))

	orm.RegisterModel(new(Account), new(Transaction), new(AccountKey))

	orm.RunSyncdb("default", false, true)
	orm.RunSyncdb("admin", false, true)
}

package pgsqldb

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/wideway/public/dev"
	myerrors "github.com/wideway/public/errors"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/wideway/public/mylog"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// MySqlDB ...
type PgSqlDB struct {
	*gorm.DB
}

// OnInitDBOrm init MySqlDB
func OnInitDBOrm(dataSourceName string, ignoreRecordNotFoundError bool) (orm *PgSqlDB) {
	orm = new(PgSqlDB)
	orm.OnGetDBOrm(dataSourceName, ignoreRecordNotFoundError)
	return
}

// OnGetDBOrm get gorm.db
func (i *PgSqlDB) OnGetDBOrm(dataSourceName string, ignoreRecordNotFoundError bool) *gorm.DB {
	if i.DB == nil {
		Default := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: ignoreRecordNotFoundError,
			Colorful:                  true,
		})
		var err error
		i.DB, err = gorm.Open(postgres.Open(dataSourceName), &gorm.Config{PrepareStmt: false,
			NamingStrategy: schema.NamingStrategy{SingularTable: true}, // 全局禁用表名复数
			Logger:         Default})                                   // logger.Default
		if err != nil {
			mylog.Error(myerrors.Wrap(err, "Got error when connect database:"+dataSourceName))
			return nil
		}
	}

	// i.DB.SingularTable(true) //全局禁用表名复数
	if dev.IsDev() {
		i.DB = i.DB.Debug()
	} else {
		i.DB.Logger = GetDBlog(ignoreRecordNotFoundError)
	}

	return i.DB
}

// OnDestoryDB destorydb
func (i *PgSqlDB) OnDestoryDB() {
	if i.DB != nil {
		sqldb, _ := i.DB.DB()
		sqldb.Close()
		i.DB = nil
	}
}

// IsNotFound 判断错误是否未找到
func (i *PgSqlDB) IsNotFound(errs ...error) bool {
	if len(errs) > 0 {
		for _, err := range errs {
			if err == gorm.ErrRecordNotFound {
				return true
			}
		}
	}
	return i.RecordNotFound()
}

// RecordNotFound check if returning ErrRecordNotFound error
func (i *PgSqlDB) RecordNotFound() bool {
	return !errors.Is(i.Error, gorm.ErrRecordNotFound)
}

// Commit 自动提交(如果有错，Rollback)
func (i *PgSqlDB) Commit(db *gorm.DB) {
	if db.Error != nil {
		db.Rollback() // 回滚
	} else {
		db.Commit()
	}
}
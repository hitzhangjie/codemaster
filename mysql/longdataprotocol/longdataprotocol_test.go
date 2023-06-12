package main_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

func TestMysqlInsert(t *testing.T) {
	// 插入普通数据
	t.Run("插入varchar数据", func(t *testing.T) {
		// 创建1个db
		db, err := openClient("root", "justdoit", 0)
		require.Nil(t, err)

		// 插入数据测试
		t.Run("normal", func(t *testing.T) {
			sqlInsertNormal := "insert into mydata(data) values(?)"
			stmt, err := db.Prepare(sqlInsertNormal)
			require.Nil(t, err)
			defer stmt.Close()

			result, err := stmt.Exec("helloworld")
			require.Nil(t, err)
			id, err := result.LastInsertId()
			require.Nil(t, err)
			require.NotZero(t, id)
		})
	})

	// 插入64MB+1B的长数据
	t.Run("longdata - 超过客户端max_allowed_packet限制(64MB)", func(t *testing.T) {
		// 创建1个db
		db, err := openClient("root", "justdoit", 0)
		require.Nil(t, err)

		db.Exec("")

		sqlInsertNormal := "insert into mydata(longdata) values(?)"
		stmt, err := db.Prepare(sqlInsertNormal)
		require.Nil(t, err)
		defer stmt.Close()

		// db server最大64MB
		_, err = stmt.Exec(makeByteSlice(1<<26 + 1))
		require.NotNil(t, err)
		merr, ok := err.(*mysql.MySQLError)
		require.True(t, ok)
		// &mysql.MySQLError{Number:0x451, Message:"Parameter of prepared statement which is set through mysql_send_long_data() is longer than 'max_allowed_packet' bytes"}
		require.Equal(t, uint16(0x451), merr.Number)
	})

	t.Run("longdata - 超过客户端")
}

func makeByteSlice(n int) []byte {
	dat := make([]byte, n)
	for i := range dat {
		dat[i] = 1
	}
	return dat
}

func openClient(user, passwd string, clientMaxAllowedPacket, serverMaxAllowedPacket int64) (*sql.DB, error) {
	var dsn string

	const defaultClientMaxAllowedPacket = 1 << 26 // 64MB
	const defaultServerMaxAllowedPacket = 1 << 26 // 64MB

	// 客户端最大包限制
	if clientMaxAllowedPacket != 0 && clientMaxAllowedPacket != defaultClientMaxAllowedPacket {
		dsn = fmt.Sprintf("%s:%s@tcp(localhost:33060)/?charset=utf8&parseTime=True&loc=Local&maxAllowedPacket=%d", user, passwd, clientMaxAllowedPacket)
	} else {
		dsn = fmt.Sprintf("%s:%s@tcp(localhost:33060)/?charset=utf8&parseTime=True&loc=Local", user, passwd)
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// 服务端最大包限制
	if serverMaxAllowedPacket != 0 && serverMaxAllowedPacket != defaultServerMaxAllowedPacket {
		// 只对后续的session生效
		db.Exec(fmt.Sprintf("set global max_allowed_packet = %d", serverMaxAllowedPacket))
		db.Close()

		// 重建session
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			return nil, err
		}
		if err = db.Ping(); err != nil {
			return nil, err
		}
	}

	// 创建db、table
	db.Exec("drop database monitor if exists")
	db.Exec("create database monitor")
	db.Exec("use monitor")
	db.Exec(`create table mydata (
		id 		int not null auto_increment,
		data 	varchar(64) not null default '',
		longdata longblob default null,
		primary key(id)
	)`)

	return db, nil
}

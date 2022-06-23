package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
)

var binlogFile = "binlog.000001"
var binlogPos = uint32(0)

func main() {
	// Create a binlog syncer with a unique server id, the server id must be different from other MySQL's.
	// flavor is mysql or mariadb
	cfg := replication.BinlogSyncerConfig{
		ServerID: 100,
		Flavor:   "mysql",
		Host:     "127.0.0.1",
		Port:     3306,
		User:     "root",
		Password: "justdoit",
	}
	syncer := replication.NewBinlogSyncer(cfg)

	// Start sync with specified binlog file and position
	streamer, _ := syncer.StartSync(mysql.Position{binlogFile, binlogPos})

	// or you can start a gtid replication like
	// streamer, _ := syncer.StartSyncGTID(gtidSet)
	// the mysql GTID set likes this "de278ad0-2106-11e4-9f8e-6edd0ca20947:1-2"
	// the mariadb GTID set likes this "0-1-100"

	for {
		ev, err := streamer.GetEvent(context.Background())
		if err != nil {
			fmt.Println("GetEvent err: %v", err)
			break
		}
		// Dump event
		ev.Dump(os.Stdout)
		vv, ok := ev.Event.(*replication.RowsEvent)
		if !ok {
			continue
		}
		b, _ := json.Marshal(vv.Table)
		fmt.Println(string(b))
	}
}

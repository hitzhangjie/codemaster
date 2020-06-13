// THIS FILE IS AUTO-GENERATED FROM fileembed.go
// DO NOT EDIT.

package testcase

import "time"

import "perkeep.org/pkg/fileembed"

func init() {
	Files.Add("fileembed.go", 116, time.Unix(0, 1592063717827760211), fileembed.String(fileembed.JoinStrings("//#fileembed pattern .+$\n"+
		"package testcase\n"+
		"\n"+
		"import (\n"+
		"	\"perkeep.org/pkg/fileembed\"\n"+
		")\n"+
		"\n"+
		"var Files = &fileembed.Files{}\n"+
		"\n"+
		"")))
}

package gen_dao

import (
	"testing"
)

func TestScannerListDir(t *testing.T) {
	//s1 := "/home/dishfo/go/src/git.uidev.tools/unifi/wire_dao/db/user_model"
	s2 := "/home/dishfo/go/src/git.uidev.tools/unifi/uid.ud/db"

	scaner, err := ScanFiles(&ScanConfig{
		TopDir: s2,
	})

	if err != nil {
		t.Fatal("scan dir failed ", err.Error())
	}

	t.Log("--- ")
	for _, dir := range scaner.dirs {
		t.Log("	---", dir)
	}

	scaner.ListAllType()
}

package domain_test

import (
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	log.Println("Setting up tests...")
	m.Run()
	log.Println("Finished testing...")
}

func TestItDoesSomething(t *testing.T) {
	t.Logf("MYSQL_URL = %s", os.Getenv("MYSQL_URL"))
}

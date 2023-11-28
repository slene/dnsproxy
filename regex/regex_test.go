package regex_test

import (
	"dnsproxy/regex"
	"fmt"
	"regexp"
	"testing"
)

func TestRegex1(t *testing.T) {
	pattern := `.*edgekey\.net$`

	// Example domains
	domains := map[string]bool{
		"edgekey.net":          true,
		"s1.edgekey.net":       true,
		"t1.s1.edgekey.net":    true,
		"aabbccedgekey.net":    true,
		"example.com":          false,
		"aabbccedgekey.net.co": false,
	}

	for domain, flag := range domains {
		match, _ := regexp.MatchString(pattern, domain)
		fmt.Printf("%s [%t]\n", domain, flag)
		if match != flag {
			t.Errorf("%s should %t, but not", domain, flag)
		}
	}
}

func TestReadHostfile(t *testing.T) {
	regex.Init("../host-file.txt")
}

func initData() map[string]string {
	hostfile := make(map[string]string)
	hostfile["*akadns.net"] = "1.1.1.1"
	hostfile["*akam.net"] = "1.1.1.2"
	hostfile["*akamai.com"] = "1.1.1.3"
	hostfile["*akamai.net"] = "1.1.1.4"
	hostfile["*akamaized.net"] = "1.1.1.5"
	hostfile["*edgekey.net"] = "1.1.1.6"
	hostfile["*edgesuite.net"] = "1.1.1.7"
	hostfile["*netflix.net"] = "1.1.1.8"
	hostfile["*netflix.com"] = "1.1.1.9"
	hostfile["*nflxext.com"] = "1.1.2.1"
	hostfile["*nflxso.net"] = "1.1.2.2"
	hostfile["*nflxso.n"] = "1.1.2.3"
	return hostfile
}

func TestConstructTree(t *testing.T) {
	hostfile := initData()

	tree := regex.ConstructTreeReverse(hostfile)
	regex.PrintTree(tree)
}

func TestFindTree(t *testing.T) {
	hostfile := initData()

	tree := regex.ConstructTreeReverse(hostfile)
	res := tree.FindReverse("netflix.com")
	if res != hostfile["*netflix.com"] {
		t.Errorf("not match [%s]", res)
	}

	res = tree.FindReverse("nnnanflxso.com")
	if res != "" {
		t.Errorf("not match [%s]", res)
	}
}

func TestFindTreeWithHostfile(t *testing.T) {
	regex.Init("../host-file.txt")

	tree := regex.MappingTree
	res := tree.FindReverse("netflix.com")
	if res != "1.1.1.1" {
		t.Errorf("not match [%s]", res)
	}
	fmt.Println(res)

	res = tree.FindReverse("nnnanflxso.net")
	if res != "1.1.1.1" {
		t.Errorf("not match [%s]", res)
	}
	fmt.Println(res)
}

package regex

import "fmt"

type Tree struct {
	Root *TreeNode
}
type TreeNode struct {
	K  rune      // key, [a]kadns.net
	B  *TreeNode // brother
	Pb *TreeNode // previous brother
	C  *TreeNode // child
	F  *TreeNode // father
	N  *Node     // value
}

type Node struct {
	V string // dns response ip, 1.1.1.1
}

func (tn *TreeNode) createFirstChild(k rune) {
	tn.C = &TreeNode{
		K: k, B: nil, Pb: nil, C: nil, F: tn, N: nil,
	}
}

func (tn *TreeNode) appendBrother(k rune) {
	if tn.B != nil {
		panic("appendBrother not nil")
	}
	tn.B = &TreeNode{
		K: k, B: nil, Pb: tn, C: nil, F: nil, N: nil,
	}
}

func (tn *TreeNode) findInBrother(k rune) (bool, *TreeNode) {
	now := tn
	prev := now
	for now != nil {
		prev = now
		if now.K == k {
			return true, now
		}
		now = now.B
	}
	return false, prev
}

func (t *Tree) construct(domain, ip string) {
	if t.Root == nil {
		t.Root = &TreeNode{}
	}
	prev := t.Root
	now := prev.C

	for _, ch := range domain {
		// fmt.Printf("%c ", ch)
		if now == nil {
			prev.createFirstChild(ch)
			prev = prev.C
			now = prev.C
			continue
		}
		f, k := now.findInBrother(ch)
		if f {
			prev = k
			now = prev.C
			continue
		} else {
			k.appendBrother(ch)
			prev = k.B
			now = prev.C
		}
	}
	prev.N = &Node{V: ip}
	// fmt.Println()
}

func reverseString(str string) string {
	runes := []rune(str)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func (t *Tree) ConstructReverse(domain, ip string) {
	t.construct(reverseString(domain), ip)
}

func ConstructTreeReverse(m map[string]string) *Tree {
	t := &Tree{}
	for k, v := range m {
		t.ConstructReverse(k, v)
	}
	return t
}

func PrintTree(t *Tree) {
	now := t.Root
	printNextChild("", now.C)
}

func printNextChild(segment string, now *TreeNode) {
	newSeg := segment + string(now.K)
	if now.N != nil {
		fmt.Printf("%s: %s\n", newSeg, now.N.V)
	}
	if now.C != nil {
		printNextChild(newSeg, now.C)
	}
	if now.B != nil {
		printNextChild(segment, now.B)
	}
}

func (t *Tree) FindReverse(domain string) (ip string) {
	domain = reverseString(domain)
	var matched bool = false
	now := t.Root.C
	for _, k := range domain {
		for now != nil {
			if now.K == k {
				matched = true
				break
			}
			if now.K == '*' {
				matched = true
				return now.N.V
			}
			now = now.B
		}
		if now != nil && now.K == k {
			now = now.C
			matched = true
			continue
		}
		matched = false
		break
	}
	if matched {
		return now.N.V
	}
	return ""
}

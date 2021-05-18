package http

import (
	"fmt"
	"github.com/dxq174510447/goframe/lib/frame/proxy"
	"sort"
	"strings"
)

type PathNode struct {
	// 目录值
	Name string
	// 匹配方式 1 绝对相同 2 单个目录配对 例如{type}这种模式 3.通用匹配* 暂时只能放在最后面 所有都没匹配上就匹配这个
	MatchType int
	// 当type为2的时候 这里设置type
	Key         string
	Children    []*PathNode
	ProxyMethod *proxy.ProxyMethod
}

func (p *PathNode) innerPrintTree(node *PathNode, depth int) {
	m := make([]string, depth, depth)
	for i := 0; i < depth; i++ {
		m[i] = "-"
	}

	fmt.Println(strings.Join(m, ""), node.Name, node.ProxyMethod)
	if len(node.Children) > 0 {
		for _, c := range node.Children {
			p.innerPrintTree(c, depth+1)
		}
	}
}
func (p *PathNode) PrintTree() {
	if len(p.Children) > 0 {
		for _, c := range p.Children {
			fmt.Println(c)
			p.innerPrintTree(c, 1)
		}
	}
}

func (p *PathNode) innerMatchMethod(current *PathNode, depth int, node []string, pathVariable map[string]string) *PathNode {
	if depth >= len(node) {
		return nil
	}

	var pathKey string
	var pathVal string
	if current.MatchType == 1 {
		if current.Name != node[depth] {
			return nil
		}
	} else if current.MatchType == 2 {
		pathKey = current.Key
		pathVal = node[depth]
	} else if current.MatchType == 3 {
		// 暂时不考虑吧
	}

	var result *PathNode
	if len(current.Children) > 0 {
		for _, c := range current.Children {
			result = p.innerMatchMethod(c, depth+1, node, pathVariable)
			if result != nil {
				break
			}
		}
	} else {
		// node第一个是空字符串
		if len(node) > (depth + 1) {
			return nil
		}
		result = current
	}

	if result != nil && pathKey != "" {
		pathVariable[pathKey] = pathVal
	}
	return result
}

// MatchMethod 最后一个匹配的node map-->{}key对应的value
func (p *PathNode) MatchMethod(requestPath string) (*PathNode, map[string]string) {
	if len(p.Children) == 0 {
		return nil, nil
	}

	r2 := requestPath
	if r2[0:1] != "/" {
		r2 = "/" + r2
	}
	if r2[len(r2)-1:len(r2)] == "/" {
		r2 = r2[0 : len(r2)-1]
	}

	r2s := strings.Split(r2, "/")
	pathVariable := make(map[string]string)
	for _, c := range p.Children {
		n := p.innerMatchMethod(c, 1, r2s, pathVariable)
		if n != nil {
			return n, pathVariable
		}
	}
	return nil, nil
}
func (p *PathNode) SetPath(configPath string, method *proxy.ProxyMethod) {
	r2 := configPath
	if r2[0:1] != "/" {
		r2 = "/" + r2
	}
	if r2[len(r2)-1:len(r2)] == "/" {
		r2 = r2[0 : len(r2)-1]
	}
	r2s := strings.Split(r2, "/")

	var currentNode *PathNode = p
	for i, k := range r2s {
		if i == 0 {
			continue
		}

		if len(currentNode.Children) > 0 {
			hasValue := false
			for _, c := range currentNode.Children {
				if c.Name == k {
					hasValue = true
					currentNode = c
					break
				}
			}
			if hasValue {
				continue
			}
		}

		var c *PathNode = &PathNode{}
		c.Name = k
		if k == "*" {
			c.MatchType = 3
		} else if k[0:1] == "{" {
			c.MatchType = 2
			c.Key = k[1 : len(k)-1]
		} else {
			c.MatchType = 1
		}
		if i == (len(r2s) - 1) {
			//last element
			c.ProxyMethod = method
		}
		currentNode.Children = append(currentNode.Children, c)
		if len(currentNode.Children) > 1 {
			sort.Slice(currentNode.Children, func(i, j int) bool {
				return currentNode.Children[i].MatchType < currentNode.Children[i].MatchType
			})
		}
		currentNode = c
	}
}

package application

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var elReg *regexp.Regexp = regexp.MustCompile(`\$\{(.+)\}`)

type YamlNode struct {
	//1 array 2 json 3 base
	NodeType  int
	NodeName  string
	NodeValue string
	RowValue  string
	ChildMap  map[string]*YamlNode
	Children  []*YamlNode
	Parent    *YamlNode
	//缩进个数
	SuojingNum int
	// 虚拟节点 数组下面如果是json 默认会创建一个虚拟节点 0否1是
	IsVirual bool
	AliasKey string
}

// YamlTree 默认根目录是空的
type YamlTree struct {
	Root    *YamlNode
	RefNode map[string]*YamlNode
	AppArgs *DefaultApplicationArguments
}

func (y *YamlTree) innerPrintTree(node *YamlNode, depth int) {
	m := make([]string, depth, depth)
	for i := 0; i < depth; i++ {
		m[i] = "-"
	}

	fmt.Println(strings.Join(m, ""), node.NodeName, node.RowValue, node.NodeType)
	if len(node.Children) > 0 {
		for _, c := range node.Children {
			y.innerPrintTree(c, depth+1)
		}
	}
}
func (y *YamlTree) PrintTree() {
	for _, c := range y.Root.Children {
		y.innerPrintTree(c, 0)
	}
}

func (y *YamlTree) getBaseValueFromNode(key string) string {

	argValue := y.AppArgs.GetByName(key, "")
	if argValue != "" {
		return argValue
	}

	// 默认从环境变量中获取
	envKey := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
	v := os.Getenv(envKey)
	if v != "" {
		return v
	}

	if y.RefNode == nil {
		return ""
	}

	k1 := fmt.Sprintf(".%s", key)

	//fmt.Println(k1)
	if current, ok := y.RefNode[k1]; ok {
		//fmt.Println("---key",k1,current.RowValue)
		return current.RowValue
	} else {
		return ""
	}
	//current := node
	//ks := strings.Split(key,".")
	//for _,k := range ks {
	//	if len(current.Children) == 0{
	//		return ""
	//	}else{
	//		if cn,ok := current.ChildMap[k] ; ok {
	//			current = cn
	//		}else{
	//			return ""
	//		}
	//	}
	//}
	//return current.RowValue
}

// GetBaseValue 暂时不支持数组元素a.b[0].c 这样中间夹杂着数组
func (y *YamlTree) GetBaseValue(key string) string {

	if y.Root == nil {
		return ""
	}

	result := y.getBaseValueFromNode(key)

	if result == "" || !isContainElexpress(result) {
		m := len(result)
		if m >= 2 && result[0:1] == "\"" && result[m-1:m] == "\"" {
			return result[1 : m-1]
		}
		return result
	}
	return y.GetElValue(result)
}

var elReg1 *regexp.Regexp = regexp.MustCompile(`\$\{`)

func (y *YamlTree) GetElValue(el1 string) string {

	str := el1
	m := elReg1.FindAllIndex([]byte(el1), -1)
	m1 := len(m) - 1
	for i := m1; i >= 0; i-- {
		if m[i][0] > 0 && str[m[i][0]-1:m[i][0]] == "\\" {
			continue
		}
		var begin int = m[i][0]
		var end int = 0
		for j := begin; j < len(str); j++ {
			if str[j:j+1] == "}" && str[j-1:j] != "\\" {
				end = j
				break
			}
		}
		if end == 0 {
			continue
		}
		// begin:end+1
		express := str[begin+2 : end]
		var key string
		var defaultValue string
		p := strings.LastIndex(express, ":")
		if p < 0 {
			key = express
			defaultValue = ""
		} else {
			key = express[0:p]
			defaultValue = express[p+1 : len(express)]
		}

		keyVal := y.GetBaseValue(key)
		if keyVal == "" {
			keyVal = defaultValue
		}
		str = fmt.Sprintf("%s%s%s", str[0:begin], keyVal, str[end+1:len(str)])
	}
	return str
}

func (y *YamlTree) setObjectValue(link string, current *YamlNode,
	rt reflect.Type,
	rv *reflect.Value) *reflect.Value {
	switch rt.Kind() {
	case reflect.Map:
		//keyrt := rt.Key()//默认是string
		if len(current.Children) == 0 {
			return nil
		}
		valrt := rt.Elem()
		for _, childNode := range current.Children {
			childkey := fmt.Sprintf("%s.%s", link, childNode.NodeName)
			switch valrt.Kind() {
			case reflect.Map:
				valvalrt := valrt.Elem()
				valmaprt := reflect.MapOf(reflect.TypeOf(""), valvalrt)
				valmaprv := reflect.MakeMap(valmaprt)
				//m1 := valmaprv.Elem()
				y.setObjectValue(childkey, childNode, valmaprt, &valmaprv)
				rv.SetMapIndex(reflect.ValueOf(childNode.NodeName), valmaprv)
			case reflect.Struct:
				valrv1 := reflect.New(valrt)
				valrv := valrv1.Elem()
				y.setObjectValue(childkey, childNode, valrt, &valrv)
				rv.SetMapIndex(reflect.ValueOf(childNode.NodeName), valrv)
			case reflect.Ptr:
				valrv1 := reflect.New(valrt.Elem())
				valrv := valrv1.Elem()
				y.setObjectValue(childkey, childNode, valrt.Elem(), &valrv)
				rv.SetMapIndex(reflect.ValueOf(childNode.NodeName), valrv1)
			case reflect.Slice:
				valarrrt := reflect.SliceOf(valrt.Elem())
				valrv1 := reflect.MakeSlice(valarrrt, 0, 0)
				//valrv := valrv1.Elem()
				sn := y.setObjectValue(childkey, childNode, valarrrt, &valrv1)
				if sn != nil {
					rv.SetMapIndex(reflect.ValueOf(childNode.NodeName), *sn)
				}
			case reflect.String:
				v := y.GetBaseValue(childkey)
				rv.SetMapIndex(reflect.ValueOf(childNode.NodeName), reflect.ValueOf(v))
			case reflect.Int:
				v := y.GetBaseValue(childkey)
				if v != "" {
					v1, _ := strconv.Atoi(v)
					rv.SetMapIndex(reflect.ValueOf(childNode.NodeName), reflect.ValueOf(v1))
				}

			case reflect.Int64:
				v := y.GetBaseValue(childkey)
				if v != "" {
					v1, _ := strconv.ParseInt(v, 10, 64)
					rv.SetMapIndex(reflect.ValueOf(childNode.NodeName), reflect.ValueOf(v1))
				}
			case reflect.Float64:
				v := y.GetBaseValue(childkey)
				if v != "" {
					v1, _ := strconv.ParseFloat(v, 64)
					rv.SetMapIndex(reflect.ValueOf(childNode.NodeName), reflect.ValueOf(v1))
				}
			}
		}
	case reflect.Struct:
		if len(current.Children) == 0 {
			return nil
		}
		nf := rt.NumField()

		var childMap map[string]*YamlNode = make(map[string]*YamlNode)
		for _, child := range current.Children {
			if child.NodeName == "" {
				continue
			}
			k := strings.ToUpper(strings.TrimSpace(removeSpecialChar(child.NodeName)))
			childMap[k] = child
		}
		for i := 0; i < nf; i++ {
			f := rt.Field(i)

			if f.Name == "" {
				continue
			}
			k := strings.ToUpper(strings.TrimSpace(removeSpecialChar(f.Name)))
			var childNode *YamlNode
			if cn, ok := childMap[k]; ok {
				childNode = cn
			} else {

				continue
			}

			childkey := fmt.Sprintf("%s.%s", link, childNode.NodeName)

			switch f.Type.Kind() {
			case reflect.Map:
				valvalrt := f.Type.Elem()
				valmaprt := reflect.MapOf(reflect.TypeOf(""), valvalrt)
				valmaprv := reflect.MakeMap(valmaprt)
				//m1 := valmaprv.Elem()
				y.setObjectValue(childkey, childNode, valmaprt, &valmaprv)
				rv.FieldByName(f.Name).Set(valmaprv)
			case reflect.Struct:
				fv := rv.FieldByName(f.Name)
				y.setObjectValue(childkey, childNode, f.Type, &fv)
			case reflect.Ptr:
				valrv1 := reflect.New(f.Type.Elem())
				valrv := valrv1.Elem()
				y.setObjectValue(childkey, childNode, f.Type.Elem(), &valrv)
				rv.FieldByName(f.Name).Set(valrv1)
			case reflect.Slice:
				valarrrt := reflect.SliceOf(f.Type.Elem())
				valrv1 := reflect.MakeSlice(valarrrt, 0, 0)

				mv := &valrv1
				sn := y.setObjectValue(childkey, childNode, valarrrt, mv)
				if sn != nil {
					rv.FieldByName(f.Name).Set(*sn)
				}
			case reflect.String:
				v := y.GetBaseValue(childkey)
				rv.FieldByName(f.Name).Set(reflect.ValueOf(v))
			case reflect.Int:
				v := y.GetBaseValue(childkey)
				if v != "" {
					v1, _ := strconv.Atoi(v)
					rv.FieldByName(f.Name).Set(reflect.ValueOf(v1))
				}

			case reflect.Int64:
				v := y.GetBaseValue(childkey)
				if v != "" {
					v1, _ := strconv.ParseInt(v, 10, 64)
					rv.FieldByName(f.Name).Set(reflect.ValueOf(v1))
				}
			case reflect.Float64:
				v := y.GetBaseValue(childkey)
				if v != "" {
					v1, _ := strconv.ParseFloat(v, 64)
					rv.FieldByName(f.Name).Set(reflect.ValueOf(v1))
				}
			}
		}
	case reflect.Ptr:
		if len(current.Children) == 0 {
			return nil
		}
		// rv不可能为空
		rv1 := rv.Elem()
		y.setObjectValue(link, current, rt.Elem(), &rv1)
	case reflect.Slice:
		var newValue *reflect.Value = rv
		if len(current.Children) == 0 {
			return nil
		}

		valrt := rt.Elem()
		for i, childNode := range current.Children {
			childkey := fmt.Sprintf("%s[%d]", link, i)
			switch valrt.Kind() {
			case reflect.Map:
				valvalrt := valrt.Elem()
				valmaprt := reflect.MapOf(reflect.TypeOf(""), valvalrt)
				valmaprv := reflect.MakeMap(valmaprt)
				y.setObjectValue(childkey, childNode, valmaprt, &valmaprv)
				sn := reflect.Append(*newValue, valmaprv)
				newValue = &sn
			case reflect.Struct:
				valrv1 := reflect.New(valrt)
				valrv := valrv1.Elem()
				y.setObjectValue(childkey, childNode, valrt, &valrv)
				sn := reflect.Append(*newValue, valrv)
				newValue = &sn
			case reflect.Ptr:
				valrv1 := reflect.New(valrt.Elem())
				valrv := valrv1.Elem()
				y.setObjectValue(childkey, childNode, valrt.Elem(), &valrv)
				sn := reflect.Append(*newValue, valrv1)
				newValue = &sn
			case reflect.Slice:
				valarrrt := reflect.SliceOf(valrt.Elem())
				valrv1 := reflect.MakeSlice(valarrrt, 0, 0)
				m1 := y.setObjectValue(childkey, childNode, valarrrt, &valrv1)
				if m1 != nil {
					sn := reflect.Append(*newValue, valrv1)
					newValue = &sn
				}

			case reflect.String:
				v := y.GetBaseValue(childkey)
				//fmt.Println(v)
				sn := reflect.Append(*newValue, reflect.ValueOf(v))
				newValue = &sn
			case reflect.Int:
				v := y.GetBaseValue(childkey)
				if v != "" {
					v1, _ := strconv.Atoi(v)
					sn := reflect.Append(*newValue, reflect.ValueOf(v1))
					newValue = &sn
				}

			case reflect.Int64:
				v := y.GetBaseValue(childkey)
				if v != "" {
					v1, _ := strconv.ParseInt(v, 10, 64)
					sn := reflect.Append(*newValue, reflect.ValueOf(v1))
					newValue = &sn
				}
			case reflect.Float64:
				v := y.GetBaseValue(childkey)
				if v != "" {
					v1, _ := strconv.ParseFloat(v, 64)
					sn := reflect.Append(*newValue, reflect.ValueOf(v1))
					newValue = &sn
				}
			}
		}
		return newValue
	}
	return nil
}

// GetObjectValue key中不支持数组ƒ√
func (y *YamlTree) GetObjectValue(key string, val interface{}) {

	var current *YamlNode

	k1 := fmt.Sprintf(".%s", key)
	if n, ok := y.RefNode[k1]; !ok {
		return
	} else {
		current = n
	}

	if current == nil || len(current.Children) == 0 {
		return
	}

	rt := reflect.TypeOf(val)
	rv := reflect.ValueOf(val)
	y.setObjectValue(key, current, rt, &rv)
}

func (y *YamlTree) reIndex1(key string, node *YamlNode, refNode map[string]*YamlNode) {
	if node == nil {
		return
	}
	if len(node.Children) == 0 {
		return
	}

	for key1, child := range node.ChildMap {
		if node.NodeType == 1 {
			var k string = fmt.Sprintf("%s[%s]", key, key1)
			refNode[k] = child
			y.reIndex1(k, child, refNode)
		} else {
			var k string = fmt.Sprintf("%s.%s", key, key1)
			refNode[k] = child
			y.reIndex1(k, child, refNode)
		}
	}
}

// ReIndex 当使用MergeTree完之后 都需要调用ReIndex重建RefNode
func (y *YamlTree) ReIndex() {
	if y.Root == nil {
		return
	}
	if len(y.Root.Children) == 0 {
		return
	}
	var refNode map[string]*YamlNode = make(map[string]*YamlNode)
	for key, child := range y.Root.ChildMap {
		key1 := fmt.Sprintf(".%s", key)
		refNode[key1] = child
		y.reIndex1(key1, child, refNode)
	}
	y.RefNode = refNode
}

func (y *YamlTree) MergeTree(target *YamlNode, source *YamlNode) {
	for k, v := range source.ChildMap {
		if tarEle, ok := target.ChildMap[k]; !ok {
			target.Children = append(target.Children, v)
			target.ChildMap[k] = v
			//y.RefNode[v.AliasKey] = v
		} else {
			if v.NodeType == 1 {
				//数组直接替换掉
				var position int = -1
				for p, c := range target.Children {
					if c.NodeName != "" && c.NodeName == k {
						position = p
						break
					}
				}
				if position < 0 {
					continue
				}
				ret := make([]*YamlNode, len(target.Children))
				copy(ret[:position], target.Children[:position])
				copy(ret[position+1:], target.Children[position+1:])
				ret[position] = v

				target.Children = ret
				target.ChildMap[k] = v

				//y.RefNode[v.AliasKey] = v
			} else {
				if len(v.Children) == 0 {
					tarEle.RowValue = v.RowValue
					if isContainElexpress(v.RowValue) {
						tarEle.NodeValue = ""
					} else {
						tarEle.NodeValue = v.RowValue
					}
				} else {
					y.MergeTree(tarEle, v)
				}
			}
		}
	}
}

// Parse 不去检验yaml格式是否正确 传入之前已经保证格式正确
func (y *YamlTree) Parse(content string) {

	if y.RefNode == nil {
		y.RefNode = make(map[string]*YamlNode)
	}

	root := &YamlNode{
		ChildMap: make(map[string]*YamlNode),
		Children: make([]*YamlNode, 0, 0),
		// 根目录默认锁进-1
		SuojingNum: -1,
	}

	reg1 := regexp.MustCompile(`(?m)^\s*$\n`)
	reg2 := regexp.MustCompile(`^\s*#`)
	m := reg1.ReplaceAllString(content, "")
	b := bufio.NewReader(bytes.NewReader([]byte(m)))

	m1, _, z1 := b.ReadLine()
	var currentNode *YamlNode = root
	for {
		if z1 != nil && z1 == io.EOF {
			break
		}
		if z1 != nil && z1 != io.EOF {
			panic(z1)
		}
		skip := false

		if reg2.Match(m1) {
			// 注释
			skip = true
		}
		if !skip {
			lineNode := y.HandleLine(string(m1), currentNode)
			if lineNode != nil {
				currentNode = lineNode
			}
		}
		m1, _, z1 = b.ReadLine()
	}

	if y.Root == nil {
		y.Root = root
	} else {
		y.MergeTree(y.Root, root)
	}
}

func (y *YamlTree) HandleLine(line string, preNode *YamlNode) *YamlNode {

	// 是否是数组
	var isArrayEle bool = false

	// 当前行缩进个数
	var suojinCount int = 0

	for i := 0; i < len(line); i++ {
		if line[i:i+1] == " " {
			suojinCount++
		} else if line[i:i+1] == "-" {
			isArrayEle = true
			suojinCount = suojinCount + 2
			break
		} else {
			break
		}
	}

	// 当前行的父节点
	var parentNode *YamlNode = nil
	if isArrayEle {
		// 数组节点
		if suojinCount > preNode.SuojingNum {
			// 如果当前的缩进大于之前一行的缩进 那上一行就是当前行的父节点
			parentNode = preNode
		} else {
			// 当前的缩进小于之前一行的缩进
			parentNode = preNode
			for {
				parentNode = parentNode.Parent
				if parentNode.SuojingNum < suojinCount {
					break
				}
			}
		}
	} else {
		if suojinCount > preNode.SuojingNum {
			// 如果当前的缩进大于之前一行的缩进 那上一行就是当前行的父节点
			parentNode = preNode
		} else {
			// 当前的缩进小于之前一行的缩进
			parentNode = preNode
			for {
				parentNode = parentNode.Parent
				if parentNode.SuojingNum < suojinCount {
					break
				}
			}
		}
	}
	if parentNode.IsVirual {
		parentNode = parentNode.Parent
	}

	if isArrayEle {
		// 设置数组节点上级节点类型
		if parentNode.NodeType == 0 {
			parentNode.NodeType = 1
			if parentNode.Children == nil {
				parentNode.Children = make([]*YamlNode, 0, 0)
				parentNode.ChildMap = make(map[string]*YamlNode)
			}
		}
	}

	var rawvalue string = line[suojinCount:len(line)]
	rawvalue = strings.TrimSpace(rawvalue)

	// 有可能没有例如数组
	position := strings.Index(rawvalue, ":")
	var rowNode *YamlNode
	if position == -1 {
		// 当前行没有 : 分隔符 只有可能是数组元素 并且值是基础类型
		rowNode = &YamlNode{}
		//1 array 2 json 3 base
		rowNode.NodeType = 3
		rowNode.NodeName = ""
		if isContainElexpress(rawvalue) {
			rowNode.NodeValue = ""
		} else {
			rowNode.NodeValue = rawvalue
		}
		rowNode.RowValue = rawvalue
		rowNode.ChildMap = nil
		rowNode.Children = nil
		rowNode.Parent = parentNode
		rowNode.SuojingNum = suojinCount

		parentNode.Children = append(parentNode.Children, rowNode)

		len1 := len(parentNode.Children) - 1
		slen := strconv.Itoa(len1)
		parentNode.ChildMap[slen] = rowNode

		rowNode.AliasKey = fmt.Sprintf("%s[%s]", parentNode.AliasKey, slen)
		y.RefNode[rowNode.AliasKey] = rowNode
	} else {
		var key string = rawvalue[0:position]
		var val string = rawvalue[position+1 : len(rawvalue)]
		key = strings.TrimSpace(key) //key  只能在空行注释
		val = strings.TrimSpace(val) //value

		//1 array 2 json 3 base
		if val != "" {
			// 当前行有key 有value

			if parentNode.NodeType == 1 {
				// array json格式
				if isArrayEle {
					virNode := &YamlNode{}
					//1 array 2 json 3 base
					virNode.NodeType = 2
					virNode.NodeName = ""
					virNode.NodeValue = ""
					virNode.RowValue = ""
					virNode.ChildMap = make(map[string]*YamlNode)
					virNode.Children = make([]*YamlNode, 0, 0)
					virNode.Parent = parentNode
					virNode.SuojingNum = suojinCount
					virNode.IsVirual = true
					parentNode.Children = append(parentNode.Children, virNode)
					len1 := len(parentNode.Children) - 1
					slen := strconv.Itoa(len1)
					parentNode.ChildMap[slen] = virNode

					virNode.AliasKey = fmt.Sprintf("%s[%s]", parentNode.AliasKey, slen)
					y.RefNode[virNode.AliasKey] = virNode

					rowNode = &YamlNode{}
					//1 array 2 json 3 base
					rowNode.NodeType = 3
					rowNode.NodeName = key
					if isContainElexpress(val) {
						rowNode.NodeValue = ""
					} else {
						rowNode.NodeValue = val
					}
					rowNode.RowValue = val
					rowNode.ChildMap = nil
					rowNode.Children = nil
					rowNode.Parent = virNode
					rowNode.SuojingNum = suojinCount
					virNode.Children = append(virNode.Children, rowNode)
					virNode.ChildMap[key] = rowNode

					rowNode.AliasKey = fmt.Sprintf("%s.%s", virNode.AliasKey, key)
					y.RefNode[rowNode.AliasKey] = rowNode
				} else {
					rowNode = &YamlNode{}
					//1 array 2 json 3 base
					rowNode.NodeType = 3
					rowNode.NodeName = key
					if isContainElexpress(val) {
						rowNode.NodeValue = ""
					} else {
						rowNode.NodeValue = val
					}
					rowNode.RowValue = val
					rowNode.ChildMap = nil
					rowNode.Children = nil

					vp := parentNode.Children[len(parentNode.Children)-1]
					rowNode.Parent = vp
					rowNode.SuojingNum = suojinCount
					vp.Children = append(vp.Children, rowNode)
					vp.ChildMap[key] = rowNode

					rowNode.AliasKey = fmt.Sprintf("%s.%s", vp.AliasKey, key)
					y.RefNode[rowNode.AliasKey] = rowNode
				}
			} else {
				//fmt.Println(line,suojinCount)
				// non array
				rowNode = &YamlNode{}
				rowNode.NodeType = 3
				rowNode.NodeName = key
				if isContainElexpress(val) {
					rowNode.NodeValue = ""
				} else {
					rowNode.NodeValue = val
				}
				rowNode.RowValue = val
				rowNode.ChildMap = nil
				rowNode.Children = nil
				rowNode.Parent = parentNode
				rowNode.SuojingNum = suojinCount

				parentNode.Children = append(parentNode.Children, rowNode)
				parentNode.ChildMap[key] = rowNode

				rowNode.AliasKey = fmt.Sprintf("%s.%s", parentNode.AliasKey, key)
				y.RefNode[rowNode.AliasKey] = rowNode
			}
		} else {
			//fmt.Println(parentNode.NodeName,parentNode.IsVirual,parentNode.NodeType,line)
			if parentNode.NodeType == 1 {
				//父节点是array
				if isArrayEle {
					//当前节点是 数组节点
					virNode := &YamlNode{}
					//1 array 2 json 3 base
					virNode.NodeType = 2
					virNode.NodeName = ""
					virNode.NodeValue = ""
					virNode.RowValue = ""
					virNode.ChildMap = make(map[string]*YamlNode)
					virNode.Children = make([]*YamlNode, 0, 0)
					virNode.Parent = parentNode
					virNode.SuojingNum = suojinCount
					virNode.IsVirual = true
					parentNode.Children = append(parentNode.Children, virNode)
					len1 := len(parentNode.Children) - 1
					slen := strconv.Itoa(len1)
					parentNode.ChildMap[slen] = virNode

					virNode.AliasKey = fmt.Sprintf("%s[%s]", parentNode.AliasKey, slen)
					y.RefNode[virNode.AliasKey] = virNode

					rowNode = &YamlNode{}
					//1 array 2 json 3 base
					//rowNode.NodeType = 3
					rowNode.NodeName = key
					rowNode.NodeValue = ""
					rowNode.RowValue = ""
					rowNode.ChildMap = make(map[string]*YamlNode)
					rowNode.Children = make([]*YamlNode, 0, 0)
					rowNode.Parent = virNode
					rowNode.SuojingNum = suojinCount
					virNode.Children = append(virNode.Children, rowNode)
					virNode.ChildMap[key] = rowNode

					rowNode.AliasKey = fmt.Sprintf("%s.%s", virNode.AliasKey, key)
					y.RefNode[rowNode.AliasKey] = rowNode
				} else {
					rowNode = &YamlNode{}
					//1 array 2 json 3 base
					//rowNode.NodeType = 3
					rowNode.NodeName = key
					rowNode.NodeValue = ""
					rowNode.RowValue = ""
					rowNode.ChildMap = make(map[string]*YamlNode)
					rowNode.Children = make([]*YamlNode, 0, 0)

					vp := parentNode.Children[len(parentNode.Children)-1]
					rowNode.Parent = vp
					rowNode.SuojingNum = suojinCount
					vp.Children = append(vp.Children, rowNode)
					vp.ChildMap[key] = rowNode

					rowNode.AliasKey = fmt.Sprintf("%s.%s", vp.AliasKey, key)
					y.RefNode[rowNode.AliasKey] = rowNode
				}
			} else {
				// rowNode.NodeType = 3
				rowNode = &YamlNode{}
				rowNode.NodeName = key
				rowNode.NodeValue = ""
				rowNode.RowValue = ""
				rowNode.ChildMap = make(map[string]*YamlNode)
				rowNode.Children = make([]*YamlNode, 0, 0)
				rowNode.Parent = parentNode
				rowNode.SuojingNum = suojinCount

				parentNode.Children = append(parentNode.Children, rowNode)
				parentNode.ChildMap[key] = rowNode

				rowNode.AliasKey = fmt.Sprintf("%s.%s", parentNode.AliasKey, key)
				y.RefNode[rowNode.AliasKey] = rowNode
			}
		}
	}
	return rowNode
}

func isContainElexpress(m string) bool {
	return elReg.Match([]byte(m))
}

var reg1 *regexp.Regexp = regexp.MustCompile(`\W+`)

func removeSpecialChar(m string) string {
	return reg1.ReplaceAllString(m, "")
}

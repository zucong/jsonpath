/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package jsonpath

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const eof = -1

const (
	leftDelim  = "{"
	rightDelim = "}"
)

type Parser struct {
	Name  string
	Root  *ListNode
	input string
	pos   int
	start int
	width int
}

var (
	ErrSyntax  = errors.New("invalid syntax")
	dictKeyRex = regexp.MustCompile(`^['"](.*)['"]$`)
	//dictKeyRex       = regexp.MustCompile(`^['"]([^']*)['"]$`)
	sliceOperatorRex = regexp.MustCompile(`^(-?[\d]*)(:-?[\d]*)?(:-?[\d]*)?$`)
)

// Parse parsed the given text and return a node Parser.
// If an error is encountered, parsing stops and an empty
// Parser is returned with the error
func Parse(name, text string) (*Parser, error) {
	p := NewParser(name)
	err := p.Parse(text) // 解析函数的入口
	if err != nil {
		p = nil
	}
	return p, err
}

func NewParser(name string) *Parser {
	return &Parser{
		Name: name,
	}
}

// parseAction parsed the expression inside delimiter
func parseAction(name, text string) (*Parser, error) {
	p, err := Parse(name, fmt.Sprintf("%s%s%s", leftDelim, text, rightDelim)) // 新建一个处理子表达式的parser, 由于parse需要大括号来作为起始和终止标志, 所以加上
	// when error happens, p will be nil, so we need to return here
	if err != nil {
		return p, err
	}
	p.Root = p.Root.Nodes[0].(*ListNode) // 由于parser会在最外面给套上一层ListNode, 所以要给这个外套脱掉, 只保留里面的实际内容
	return p, nil
}

func (p *Parser) Parse(text string) error {
	p.input = text
	p.Root = newList()
	p.pos = 0
	return p.parseText(p.Root)
}

// consumeText return the parsed text since last cosumeText
func (p *Parser) consumeText() string {
	value := p.input[p.start:p.pos]
	p.start = p.pos
	return value
}

// next returns the next rune in the input.
func (p *Parser) next() rune {
	if p.pos >= len(p.input) {
		p.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(p.input[p.pos:])
	p.width = w
	p.pos += p.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (p *Parser) peek() rune {
	r := p.next()
	p.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (p *Parser) backup() {
	p.pos -= p.width
}

func (p *Parser) parseText(cur *ListNode) error {
	for {
		if strings.HasPrefix(p.input[p.pos:], leftDelim) {
			if p.pos > p.start { // 从起始字符到左括号之间有若干个普通字符, 先把它做成个TextNode放到当前Node列表里, 再处理左括号和后续的东西
				cur.append(newText(p.consumeText()))
			}
			return p.parseLeftDelim(cur)
		}
		if p.next() == eof {
			break
		}
	}
	// Correctly reached EOF.
	if p.pos > p.start {
		cur.append(newText(p.consumeText()))
	}
	return nil
}

// parseLeftDelim scans the left delimiter, which is known to be present.
func (p *Parser) parseLeftDelim(cur *ListNode) error {
	p.pos += len(leftDelim)
	p.consumeText()                 // 直接消耗掉这个左大括号
	newNode := newList()            // 大括号里面这些东西的形式是不固定的(可能很多也可能很少), 所以要new一个ListNode来存放里面的若干Node
	cur.append(newNode)             // 把这个ListNode放到上层ListNode列表里面
	cur = newNode                   // 然后cur指向了当前层次的ListNode
	return p.parseInsideAction(cur) // 进行大括号内部的解析
}

func (p *Parser) parseInsideAction(cur *ListNode) error {
	prefixMap := map[string]func(*ListNode) error{ // 大括号里面可能会有这三种特殊情况, 这些要另开个新的处理流程
		rightDelim: p.parseRightDelim,
		"[?(":      p.parseFilter,
		"..":       p.parseRecursive,
	}
	for prefix, parseFunc := range prefixMap { // 看一看到底是哪一种特殊情况, 用对应的解析方法来处理
		if strings.HasPrefix(p.input[p.pos:], prefix) {
			return parseFunc(cur)
		}
	}

	switch r := p.next(); { // 非特殊情况的处理
	case r == eof || isEndOfLine(r):
		return fmt.Errorf("unclosed action")
	case r == ' ': // 遇到空格直接消耗掉
		p.consumeText()
	case r == '@' || r == '$': // 这种字符代表当前的对象, 直接消耗掉, 然后递归后续表达式处理流程
		p.consumeText()
	case r == '[':
		return p.parseArray(cur)
	case r == '"' || r == '\'':
		return p.parseQuote(cur, r)
	case r == '.':
		return p.parseField(cur)
	case r == '+' || r == '-' || unicode.IsDigit(r):
		p.backup()
		return p.parseNumber(cur)
	case isAlphaNumeric(r):
		p.backup()
		return p.parseIdentifier(cur)
	default:
		return fmt.Errorf("unrecognized character in action: %#U", r)
	}
	return p.parseInsideAction(cur) // 递归处理后续字符串
}

// parseRightDelim scans the right delimiter, which is known to be present.
func (p *Parser) parseRightDelim(cur *ListNode) error { // 遇到右大括号表示处理的结束
	p.pos += len(rightDelim)
	p.consumeText()
	return p.parseText(p.Root) // 看一下右大括号后面还有没有别的东西
}

// parseIdentifier scans build-in keywords, like "range" "end"
func (p *Parser) parseIdentifier(cur *ListNode) error {
	var r rune
	for {
		r = p.next()
		if isTerminator(r) {
			p.backup()
			break
		}
	}
	value := p.consumeText()

	if isBool(value) {
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("can not parse bool '%s': %s", value, err.Error())
		}

		cur.append(newBool(v))
	} else {
		cur.append(newIdentifier(value))
	}

	return p.parseInsideAction(cur)
}

// parseRecursive scans the recursive descent operator ..
func (p *Parser) parseRecursive(cur *ListNode) error {
	if lastIndex := len(cur.Nodes) - 1; lastIndex >= 0 && cur.Nodes[lastIndex].Type() == NodeRecursive {
		return fmt.Errorf("invalid multiple recursive descent")
	}
	p.pos += len("..")
	p.consumeText()
	cur.append(newRecursive())
	if r := p.peek(); isAlphaNumeric(r) {
		return p.parseField(cur)
	}
	return p.parseInsideAction(cur)
}

// parseNumber scans number
func (p *Parser) parseNumber(cur *ListNode) error {
	r := p.peek()
	if r == '+' || r == '-' {
		p.next()
	}
	for {
		r = p.next()
		if r != '.' && !unicode.IsDigit(r) {
			p.backup()
			break
		}
	}
	value := p.consumeText()
	i, err := strconv.Atoi(value)
	if err == nil {
		cur.append(newInt(i))
		return p.parseInsideAction(cur)
	}
	d, err := strconv.ParseFloat(value, 64)
	if err == nil {
		cur.append(newFloat(d))
		return p.parseInsideAction(cur)
	}
	return fmt.Errorf("cannot parse number %s", value)
}

func (p *Parser) findNextRune(r rune, cur *ListNode) error {
	c := rune(0)
	escapeMode := false
	for {
		c = p.next()
		if c == r && !escapeMode {
			return nil
		} else if c == '\\' && !escapeMode {
			escapeMode = true
		} else if c == eof {
			return fmt.Errorf("cannot find the next %c", r)
		} else {
			escapeMode = false
		}
	}
}

func splitByComma(str string) []string {
	result := make([]string, 0)
	base := 0
	next := -1
	rs := []rune(str)
	for i := 0; i < len(rs); i++ {
		if rs[i] == ',' {
			result = append(result, string(rs[base:i]))
			base = i + 1
		} else if rs[i] == '\'' || rs[i] == '"' {
			next = findRune(rs[i+1:], rs[i])
			if next == -1 {
				return nil
			} else {
				i += next + 1
			}
		}
	}
	result = append(result, string(rs[base:]))
	return result
}

func findRune(rs []rune, target rune) int {
	escapeMode := false
	for i, r := range rs {
		if r == target && !escapeMode {
			return i
		} else if r == '\\' && !escapeMode {
			escapeMode = true
		} else if r == eof {
			return -1
		} else {
			escapeMode = false
		}
	}
	return -1
}

// parseArray scans array index selection
func (p *Parser) parseArray(cur *ListNode) error {
Loop:
	for {
		r := p.next()
		switch r {
		case eof, '\n':
			return fmt.Errorf("unterminated array")
		case '"':
			fallthrough
		case '\'':
			err := p.findNextRune(r, cur)
			if err != nil {
				return err
			}
		case ']':
			break Loop
		}
	}
	text := p.consumeText()
	text = text[1 : len(text)-1]
	if text == "*" {
		//text = ":"
		cur.append(newWildcard())
		return p.parseInsideAction(cur)
	}

	//union operator
	//strs := strings.Split(text, ",")
	strs := splitByComma(text)
	if len(strs) > 1 {
		union := []*ListNode{}
		for _, str := range strs {
			parser, err := parseAction("union", fmt.Sprintf("[%s]", strings.Trim(str, " ")))
			if err != nil {
				return err
			}
			union = append(union, parser.Root)
		}
		cur.append(newUnion(union))
		return p.parseInsideAction(cur)
	}

	// dict key
	text = strings.TrimSpace(text)
	value := dictKeyRex.FindStringSubmatch(text)
	if value != nil {
		//parser, err := parseAction("arraydict", fmt.Sprintf(".%s", value[1]))
		//if err != nil {
		//	return err
		//}
		//for _, node := range parser.Root.Nodes {
		//	cur.append(node)
		//}
		cur.append(newField(value[1]))
		return p.parseInsideAction(cur)
	}

	//slice operator
	value = sliceOperatorRex.FindStringSubmatch(text)
	if value == nil {
		return fmt.Errorf("invalid array index %s", text)
	}
	value = value[1:]
	if value[1] == "" && value[2] == "" {
		var arrayElement *ArrayElementNode
		if value[0] == "" {
			arrayElement = newArrayElement(ParamsEntry{
				Value:   0,
				Known:   false,
				Derived: false,
			})
		} else {
			i, err := strconv.Atoi(value[0])
			if err != nil {
				return fmt.Errorf("array index %s is not a number", value[i])
			}
			arrayElement = newArrayElement(ParamsEntry{
				Value:   i,
				Known:   true,
				Derived: false,
			})
		}
		cur.append(arrayElement)
		return p.parseInsideAction(cur)
	}
	params := make([]ParamsEntry, 3)
	for i := 0; i < 3; i++ {
		if value[i] != "" {
			if i > 0 {
				value[i] = value[i][1:]
			}
			if i > 0 && value[i] == "" {
				params[i].Known = false
			} else {
				var err error
				params[i].Known = true
				params[i].Value, err = strconv.Atoi(value[i])
				if err != nil {
					return fmt.Errorf("array index %s is not a number", value[i])
				}
			}
		} else {
			params[i].Known = false
			params[i].Value = 0
		}
	}
	cur.append(newArray(params))
	return p.parseInsideAction(cur)
}

// parseFilter scans filter inside array selection
func (p *Parser) parseFilter(cur *ListNode) error {
	p.pos += len("[?(")
	p.consumeText() // 消耗掉这个[?(
	begin := false
	end := false
	var pair rune

Loop:
	for {
		r := p.next()
		switch r {
		case eof, '\n': // filter里面不能有这种东西, 否则乱套了, 报错返回
			return fmt.Errorf("unterminated filter")
		case '"', '\'': // 双引号和单引号都是是要成对出现的
			if begin == false {
				//save the paired rune
				begin = true
				pair = r
				continue
			}
			//only add when met paired rune
			if p.input[p.pos-2] != '\\' && r == pair {
				end = true
			}
		case ')': // 代表filter结束了, 这个右小括号只能出现一次
			//in rightParser below quotes only appear zero or once
			//and must be paired at the beginning and end
			if begin == end {
				break Loop
			}
		}
	}
	if p.next() != ']' {
		return fmt.Errorf("unclosed array expect ]")
	}
	reg := regexp.MustCompile(`^([^!<>=]+)([!<>=]+)(.+?)$`)
	text := p.consumeText()
	text = text[:len(text)-2]             // 提取出整个filter字符串
	value := reg.FindStringSubmatch(text) // 把filter字符串按照正则表达式里的小括号切分成三个部分: "引用(左表达式)", "符号", "字面值(右表达式)"
	if value == nil {
		parser, err := parseAction("text", text)
		if err != nil {
			return err
		}
		cur.append(newFilter(parser.Root, newList(), "exists"))
	} else {
		leftParser, err := parseAction("left", value[1]) // 子parser, 包含了左表达式里的Nodes
		if err != nil {
			return err
		}
		rightParser, err := parseAction("right", value[3])
		if err != nil {
			return err
		}
		cur.append(newFilter(leftParser.Root, rightParser.Root, value[2]))
	}
	return p.parseInsideAction(cur)
}

// parseQuote unquotes string inside double or single quote
func (p *Parser) parseQuote(cur *ListNode, end rune) error { // 处理引号
Loop:
	for {
		switch p.next() {
		case eof, '\n':
			return fmt.Errorf("unterminated quoted string")
		case end:
			//if it's not escape break the Loop
			if p.input[p.pos-2] != '\\' {
				break Loop
			}
		}
	}
	value := p.consumeText()       // 取出整个引号字符串
	s, err := UnquoteExtend(value) // 去掉引号
	if err != nil {
		return fmt.Errorf("unquote string %s error %v", value, err)
	}
	cur.append(newText(s))
	return p.parseInsideAction(cur)
}

// parseField scans a field until a terminator
func (p *Parser) parseField(cur *ListNode) error { // 处理属性成员类型
	p.consumeText() // 先消耗掉这个'.'
	for p.advance() {
	}
	value := p.consumeText() // 把属性成员的名字消耗掉, 把名字进行下面的处理
	if value == "*" {        // 如果名字是个通配符
		cur.append(newWildcard())
	} else { // 普通名字
		cur.append(newField(strings.Replace(value, "\\", "", -1)))
	}
	return p.parseInsideAction(cur) // 处理后续东西
}

// advance scans until next non-escaped terminator
func (p *Parser) advance() bool { // 前进知道遇到了分隔符
	r := p.next()
	if r == '\\' {
		p.next()
	} else if isTerminator(r) {
		p.backup()
		return false
	}
	return true
}

// isTerminator reports whether the input is at valid termination character to appear after an identifier.
func isTerminator(r rune) bool { // 判断是否遇到了分隔符
	if isSpace(r) || isEndOfLine(r) {
		return true
	}
	switch r {
	case eof, '.', ',', '[', ']', '$', '@', '{', '}':
		return true
	}
	return false
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// isBool reports whether s is a boolean value.
func isBool(s string) bool {
	return s == "true" || s == "false"
}

//UnquoteExtend is almost same as strconv.Unquote(), but it support parse single quotes as a string
func UnquoteExtend(s string) (string, error) {
	n := len(s)
	if n < 2 {
		return "", ErrSyntax
	}
	quote := s[0]
	if quote != s[n-1] {
		return "", ErrSyntax
	}
	s = s[1 : n-1]

	if quote != '"' && quote != '\'' {
		return "", ErrSyntax
	}

	// Is it trivial?  Avoid allocation.
	if !contains(s, '\\') && !contains(s, quote) {
		return s, nil
	}

	var runeTmp [utf8.UTFMax]byte
	buf := make([]byte, 0, 3*len(s)/2) // Try to avoid more allocations.
	for len(s) > 0 {
		c, multibyte, ss, err := strconv.UnquoteChar(s, quote)
		if err != nil {
			return "", err
		}
		s = ss
		if c < utf8.RuneSelf || !multibyte {
			buf = append(buf, byte(c))
		} else {
			n := utf8.EncodeRune(runeTmp[:], c)
			buf = append(buf, runeTmp[:n]...)
		}
	}
	return string(buf), nil
}

func contains(s string, c byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return true
		}
	}
	return false
}

package resource

import (
	"fmt"

	"github.com/qor/qor"
)

type callback struct {
	creates    []*func(context *qor.Context)
	updates    []*func(context *qor.Context)
	deletes    []*func(context *qor.Context)
	queries    []*func(context *qor.Context)
	processors []*callbackProcessor
}

type callbackProcessor struct {
	name      string
	before    string
	after     string
	replace   bool
	remove    bool
	typ       string
	processor *func(context *qor.Context)
	callback  *callback
}

func (c *callback) addProcessor(typ string) *callbackProcessor {
	cp := &callbackProcessor{typ: typ, callback: c}
	c.processors = append(c.processors, cp)
	return cp
}

func (c *callback) clone() *callback {
	return &callback{
		creates:    c.creates,
		updates:    c.updates,
		deletes:    c.deletes,
		queries:    c.queries,
		processors: c.processors,
	}
}

func (c *callback) Create() *callbackProcessor {
	return c.addProcessor("create")
}

func (c *callback) Update() *callbackProcessor {
	return c.addProcessor("update")
}

func (c *callback) Delete() *callbackProcessor {
	return c.addProcessor("delete")
}

func (c *callback) Query() *callbackProcessor {
	return c.addProcessor("query")
}

func (cp *callbackProcessor) Before(name string) *callbackProcessor {
	cp.before = name
	return cp
}

func (cp *callbackProcessor) After(name string) *callbackProcessor {
	cp.after = name
	return cp
}

func (cp *callbackProcessor) Register(name string, fc func(context *qor.Context)) {
	cp.name = name
	cp.processor = &fc
	cp.callback.sort()
}

func (cp *callbackProcessor) Remove(name string) {
	fmt.Printf("[info] removing callback `%v` from %v\n", name, qor.FilenameWithLineNum())
	cp.name = name
	cp.remove = true
	cp.callback.sort()
}

func (cp *callbackProcessor) Replace(name string, fc func(context *qor.Context)) {
	fmt.Printf("[info] replacing callback `%v` from %v\n", name, qor.FilenameWithLineNum())
	cp.name = name
	cp.processor = &fc
	cp.replace = true
	cp.callback.sort()
}

func getRIndex(strs []string, str string) int {
	for i := len(strs) - 1; i >= 0; i-- {
		if strs[i] == str {
			return i
		}
	}
	return -1
}

func sortProcessors(cps []*callbackProcessor) []*func(context *qor.Context) {
	var sortCallbackProcessor func(c *callbackProcessor)
	var names, sortedNames = []string{}, []string{}

	for _, cp := range cps {
		if index := getRIndex(names, cp.name); index > -1 {
			if !cp.replace && !cp.remove {
				fmt.Printf("[warning] duplicated callback `%v` from %v\n", cp.name, qor.FilenameWithLineNum())
			}
		}
		names = append(names, cp.name)
	}

	sortCallbackProcessor = func(c *callbackProcessor) {
		if getRIndex(sortedNames, c.name) > -1 {
			return
		}

		if len(c.before) > 0 {
			if index := getRIndex(sortedNames, c.before); index > -1 {
				sortedNames = append(sortedNames[:index], append([]string{c.name}, sortedNames[index:]...)...)
			} else if index := getRIndex(names, c.before); index > -1 {
				sortedNames = append(sortedNames, c.name)
				sortCallbackProcessor(cps[index])
			} else {
				sortedNames = append(sortedNames, c.name)
			}
		}

		if len(c.after) > 0 {
			if index := getRIndex(sortedNames, c.after); index > -1 {
				sortedNames = append(sortedNames[:index+1], append([]string{c.name}, sortedNames[index+1:]...)...)
			} else if index := getRIndex(names, c.after); index > -1 {
				cp := cps[index]
				if len(cp.before) == 0 {
					cp.before = c.name
				}
				sortCallbackProcessor(cp)
			} else {
				sortedNames = append(sortedNames, c.name)
			}
		}

		if getRIndex(sortedNames, c.name) == -1 {
			sortedNames = append(sortedNames, c.name)
		}
	}

	for _, cp := range cps {
		sortCallbackProcessor(cp)
	}

	var funcs = []*func(context *qor.Context){}
	var sortedFuncs = []*func(context *qor.Context){}
	for _, name := range sortedNames {
		index := getRIndex(names, name)
		if !cps[index].remove {
			sortedFuncs = append(sortedFuncs, cps[index].processor)
		}
	}

	for _, cp := range cps {
		if sindex := getRIndex(sortedNames, cp.name); sindex == -1 {
			if !cp.remove {
				funcs = append(funcs, cp.processor)
			}
		}
	}

	return append(sortedFuncs, funcs...)
}

func (c *callback) sort() {
	creates, updates, deletes, queries := []*callbackProcessor{}, []*callbackProcessor{}, []*callbackProcessor{}, []*callbackProcessor{}

	for _, processor := range c.processors {
		switch processor.typ {
		case "create":
			creates = append(creates, processor)
		case "update":
			updates = append(updates, processor)
		case "delete":
			deletes = append(deletes, processor)
		case "query":
			queries = append(queries, processor)
		}
	}

	c.creates = sortProcessors(creates)
	c.updates = sortProcessors(updates)
	c.deletes = sortProcessors(deletes)
	c.queries = sortProcessors(queries)
}

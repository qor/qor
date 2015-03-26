package resource

import (
	"fmt"

	"github.com/qor/qor"
)

type Callback struct {
	creates    []*func(context *qor.Context) error
	updates    []*func(context *qor.Context) error
	deletes    []*func(context *qor.Context) error
	queries    []*func(context *qor.Context) error
	processors []*CallbackProcessor
}

type CallbackProcessor struct {
	name      string
	before    string
	after     string
	replace   bool
	remove    bool
	typ       string
	processor *func(context *qor.Context) error
	callback  *Callback
}

func (c *Callback) addProcessor(typ string) *CallbackProcessor {
	cp := &CallbackProcessor{typ: typ, callback: c}
	c.processors = append(c.processors, cp)
	return cp
}

func (c *Callback) clone() *Callback {
	return &Callback{
		creates:    c.creates,
		updates:    c.updates,
		deletes:    c.deletes,
		queries:    c.queries,
		processors: c.processors,
	}
}

func (c *Callback) Create() *CallbackProcessor {
	return c.addProcessor("create")
}

func (c *Callback) Update() *CallbackProcessor {
	return c.addProcessor("update")
}

func (c *Callback) Delete() *CallbackProcessor {
	return c.addProcessor("delete")
}

func (c *Callback) Query() *CallbackProcessor {
	return c.addProcessor("query")
}

func (cp *CallbackProcessor) Before(name string) *CallbackProcessor {
	cp.before = name
	return cp
}

func (cp *CallbackProcessor) After(name string) *CallbackProcessor {
	cp.after = name
	return cp
}

func (cp *CallbackProcessor) Register(name string, fc func(context *qor.Context) error) {
	cp.name = name
	cp.processor = &fc
	cp.callback.sort()
}

func (cp *CallbackProcessor) Remove(name string) {
	fmt.Printf("[info] removing callback `%v` from %v\n", name, qor.FilenameWithLineNum())
	cp.name = name
	cp.remove = true
	cp.callback.sort()
}

func (cp *CallbackProcessor) Replace(name string, fc func(context *qor.Context) error) {
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

func sortProcessors(cps []*CallbackProcessor) []*func(context *qor.Context) error {
	var sortCallbackProcessor func(c *CallbackProcessor)
	var names, sortedNames = []string{}, []string{}

	for _, cp := range cps {
		if index := getRIndex(names, cp.name); index > -1 {
			if !cp.replace && !cp.remove {
				fmt.Printf("[warning] duplicated callback `%v` from %v\n", cp.name, qor.FilenameWithLineNum())
			}
		}
		names = append(names, cp.name)
	}

	sortCallbackProcessor = func(c *CallbackProcessor) {
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

	var funcs = []*func(*qor.Context) error{}
	var sortedFuncs = []*func(*qor.Context) error{}
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

func (c *Callback) sort() {
	creates, updates, deletes, queries := []*CallbackProcessor{}, []*CallbackProcessor{}, []*CallbackProcessor{}, []*CallbackProcessor{}

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

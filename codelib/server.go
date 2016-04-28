// Copyright 2016 ArkBriar. All rights reserved.
package codelib

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
)

type Command interface {
	Command() string                 // return the name of this command
	Execute(map[string]string) error // execute this command
	Help(io.Writer)                  // print the help infomation
	Brief() string                   // brief introduction
}

type (
	FilterFunc  func(command string) (map[string]string, error)
	ExecuteFunc func(map[string]string) error
	HelpFunc    func(io.Writer)
)

type __CommandImpl struct {
	CommandName string

	// functions
	execute ExecuteFunc
	help    HelpFunc
	filter  FilterFunc
	brief   func() string
}

func (this *__CommandImpl) Command() string {
	return this.CommandName
}

func (this *__CommandImpl) Execute(argv map[string]string) error {
	return this.execute(argv)
}

func (this *__CommandImpl) Help(w io.Writer) {
	this.help(w)
}

func (this *__CommandImpl) Brief() string {
	if this.brief != nil {
		return this.brief()
	}
	return this.CommandName
}

type CommandEntity struct {
	command Command
	argv    map[string]string
}

func (this *CommandEntity) Execute() error {
	return this.command.Execute(this.argv)
}

// we can use command registered here
var __CommandMap = make(map[string]*__CommandImpl)

func RegisterCommand(name string, execute ExecuteFunc, filter FilterFunc, help HelpFunc, brief func() string) error {
	if _, ok := __CommandMap[name]; ok {
		return errors.New("Command " + name + " is already registed!")
	}

	__CommandMap[name] = &__CommandImpl{
		CommandName: name,
		execute:     execute,
		help:        help,
		filter:      filter,
		brief:       brief,
	}

	return nil
}

func __GetCommand(name string) (*__CommandImpl, bool) {
	cimpl, ok := __CommandMap[name]
	return cimpl, ok
}

func ListCommands(w io.Writer) {
	fmt.Fprintln(w, "There are "+strconv.Itoa(len(__CommandMap))+" commands:")
	for cname, cimpl := range __CommandMap {
		fmt.Fprintln(w, "\t"+cname+": "+cimpl.Brief())
	}
}

func __Split(command string) (string, string) {
	t := bytes.SplitN([]byte(command), []byte(" "), 2)
	return string(t[0]), string(t[1])
}

func NewCommand(command string) (*CommandEntity, error) {
	fmt.Println()

	name, argv_s := __Split(command)

	if c, ok := __GetCommand(name); !ok {
		return nil, errors.New("There's no command called " + c.Command())
	} else {
		argv, err := c.filter(argv_s)
		if err != nil {
			return nil, err
		}

		return &CommandEntity{
			command: c,
			argv:    argv,
		}, nil
	}
}

func ReadLine() string {
	//@TODO
	return ""
}

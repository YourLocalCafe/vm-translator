package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)


type Parser struct {
	file *os.File
	scanner *bufio.Reader
	CurrentInstruction string
}


func NewParser(filename string) (*Parser, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	sc := bufio.NewReader(f)
	return &Parser{
		file: f,
		scanner: sc,
		CurrentInstruction: "",
	}, nil
}


func (p *Parser) Advance() (bool, error) {
	hasNext := true
	for {
		rawLine, err := p.scanner.ReadString('\n')
		switch err {
		case nil:
		case io.EOF:
			hasNext = false
		default:
			return false, err
		}
		command, _, hasComment := strings.Cut(strings.Trim(rawLine, " \n\r"), "//")
		if (hasComment || len(command) == 0) && hasNext {
			continue
		}
		p.CurrentInstruction = command
		return hasNext, nil
	}
}


func (p *Parser) CommandType() (string, error) {
	switch commandPrefix, _, _ := strings.Cut(p.CurrentInstruction, " "); commandPrefix {
	case "add", "sub", "lt", "gt", "eq", "and", "or", "neg", "not":
		return "C_ARITHMETIC", nil
	case "pop":
		return "C_POP", nil
	case "push":
		return "C_PUSH", nil
	default:
		errorMessage := fmt.Sprintf("Invalid command type with prefix: %s\n", commandPrefix)
		return "", errors.New(errorMessage)
	}
}


func (p *Parser) Arg1() (string, error) {
	commandType, err := p.CommandType()
	if err != nil {
		return "", err
	}
	switch commandType {
	case "C_ARITHMETIC":
		return p.CurrentInstruction, nil
	case "C_PUSH", "C_POP":
		return strings.Fields(p.CurrentInstruction)[1], nil
	default:
		errorMessage := fmt.Sprintf("Invalid command type: %s\n", commandType)
		return "", errors.New(errorMessage)
	}
}


func (p *Parser) Arg2() (int, error) {
	argTwo, err := strconv.Atoi(strings.Fields(p.CurrentInstruction)[2])
	if err != nil {
		return -1, err
	}
	return argTwo, nil
}
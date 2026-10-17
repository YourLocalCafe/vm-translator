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
	LabelMap map[string]map[string]struct{}
	CurrentFunction []string
}


func NewParser(filename string) (*Parser, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	sc := bufio.NewReader(f)
	functionName := []string{"dummyFunctionWrapper"}
	labelMapping := map[string]map[string]struct{}{}
	for {
		rawLine, err := sc.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		commandWithoutComment, _, _ := strings.Cut(rawLine, "//")
		command := strings.Trim(commandWithoutComment, " \t\n\r")
		if len(command) == 0 {
			continue
		}
		commandParts := strings.Fields(command)
		switch commandParts[0] {
		case "label":
			if _, exists := labelMapping[commandParts[1]]; !exists {
				labelMapping[commandParts[1]] = map[string]struct{}{}
			}
			if len(functionName) > 0 {
				labelMapping[commandParts[1]][functionName[len(functionName) - 1]] = struct{}{}
			} else {
				errorMessage := fmt.Sprintf("An error occurred during parsing - Invalid use of \"label\", function call stack: %s", functionName)
				return nil, errors.New(errorMessage)
			}
		case "function":
			functionName = append(functionName, commandParts[1])
		case "return":
			if len(functionName) > 0 {
				functionName = functionName[:len(functionName) - 1]
			} else {
				errorMessage := fmt.Sprintf("An error occurred during parsing - Invalid use of return outside function, function call stack: %s", functionName)
				return nil, errors.New(errorMessage)
			}
		}
	}
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}
	return &Parser{
		file: f,
		scanner: sc,
		CurrentInstruction: "",
		LabelMap: labelMapping,
		CurrentFunction: []string{"dummyFunctionWrapper"},
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
		commandWithoutComment, _, _ := strings.Cut(rawLine, "//")
		command := strings.Trim(commandWithoutComment, " \t\n\r")
		if len(command) == 0 && hasNext {
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
	case "label":
		return "C_LABEL", nil
	case "goto":
		labelName, err := p.Arg1()
		_, exists := p.LabelMap[labelName]
		if !exists {
			errorMessage := fmt.Sprintf("An error occurred during parsing - Invalid label supplied to goto instruction: %s", labelName)
			return "", errors.New(errorMessage)
		}
		found := false
		for i := range p.LabelMap[labelName] {
			if p.CurrentFunction[len(p.CurrentFunction) - 1] == i {
				found = true
				break
			}
		}
		if !found {
			errorMessage := fmt.Sprintf("An error occurred during parsing - Label supplied in goto instruction not within function scope: %s", labelName)
			return "", errors.New(errorMessage)
		}
		return "C_GOTO", err
	case "if-goto":
		labelName, err := p.Arg1()
		_, exists := p.LabelMap[labelName]
		if !exists {
			errorMessage := fmt.Sprintf("An error occurred during parsing - Invalid label supplied to if-goto instruction: %s", labelName)
			return "", errors.New(errorMessage)
		}
		found := false
		for i := range p.LabelMap[labelName] {
			if p.CurrentFunction[len(p.CurrentFunction) - 1] == i {
				found = true
				break
			}
		}
		if !found {
			errorMessage := fmt.Sprintf("An error occurred during parsing - Label supplied in if-goto instruction not within function scope: %s", labelName)
			return "", errors.New(errorMessage)
		}
		return "C_IF", err
	case "function":
		functionName, err := p.Arg1()
		p.CurrentFunction = append(p.CurrentFunction, functionName)
		return "C_FUNCTION", err
	case "call":
		return "C_CALL", nil
	case "return":
		if len(p.CurrentFunction) > 0 {
			p.CurrentFunction = p.CurrentFunction[:len(p.CurrentFunction) - 1]
		} else {
			errorMessage := fmt.Sprintf("An error occurred during parsing - Invalid use of return outside function, function call stack: %s", p.CurrentFunction)
			return "", errors.New(errorMessage)
		}
		return "C_RETURN", nil
	default:
		errorMessage := fmt.Sprintf("An error occurred during parsing - Invalid command type with prefix: %s", commandPrefix)
		return "", errors.New(errorMessage)
	}
}


func (p *Parser) Arg1() (string, error) {
	commandParts := strings.Fields(p.CurrentInstruction)
	if len(commandParts) == 1 && p.CurrentInstruction == "return" {
		return "", errors.New("An error occurred during parsing - \"return\" has no arguments")
	} else if len(commandParts) == 1 {
		return p.CurrentInstruction, nil
	} else {
		return commandParts[1], nil
	}
}


func (p *Parser) Arg2() (int, error) {
	argTwo, err := strconv.Atoi(strings.Fields(p.CurrentInstruction)[2])
	if err != nil {
		return -1, err
	}
	return argTwo, nil
}


func (p *Parser) Cleanup() (error) {
	err := p.file.Close()
	return err
}
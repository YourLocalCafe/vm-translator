package main

import (
	"fmt"
	"os"
	"strings"
	"unicode"
	"github.com/YourLocalCafe/vm-translator/parser"
	"github.com/YourLocalCafe/vm-translator/writer"
)


func translate(fileName string) error {
	p, err := parser.NewParser(fileName)
	if err != nil {
		return err
	}
	defer p.Cleanup()

	fileNameWithoutExtension, _ := strings.CutSuffix(fileName, ".vm")
	w, err := writer.NewWriter(fileNameWithoutExtension + ".asm")
	if err != nil {
		return err
	}
	defer w.Close()

	for {
		hasNext, err := p.Advance()
		if err != nil {
			return err
		}
		if !hasNext && len(p.CurrentInstruction) == 0 {
			break
		}

		commandType, err := p.CommandType()
		if err != nil {
			return err
		}

		var arg1 string
		if commandType != "C_RETURN" {
			arg1, err = p.Arg1()
			if err != nil {
				return err
			}
		}
		var arg2 int
		switch commandType {
		case "C_PUSH", "C_POP", "C_FUNCTION", "C_CALL":
			arg2, err = p.Arg2()
			if err != nil {
				return err
			}
		}

		switch commandType {
		case "C_ARITHMETIC":
			err = w.WriteArithmetic(p.CurrentInstruction)
			if err != nil {
				return err
			}
		case "C_PUSH", "C_POP":
			err = w.WritePushPop(commandType, arg1, arg2)
			if err != nil {
				return err
			}
		}
	}

	return nil
}


func main() {
	args := os.Args
	if len(args) != 2 {
		fmt.Println("Usage: ./vm-translator <Filename>.vm")
		os.Exit(1)
	}
	fileName := args[1]
	filePath := strings.Split(fileName, "/")
	if !unicode.IsUpper(rune(filePath[len(filePath) - 1][0])) || !strings.HasSuffix(fileName, ".vm") {
		fmt.Println("Invalid filename - should be of the format: Filename.vm")
		os.Exit(1)
	}

	err := translate(fileName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
package main

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/YourLocalCafe/vm-translator/parser"
	"github.com/YourLocalCafe/vm-translator/writer"
)


func translate(fileName string, w *writer.Writer) error {
	p, err := parser.NewParser(fileName)
	if err != nil {
		return err
	}
	defer p.Cleanup()

	fileNameWithoutExtension, _ := strings.CutSuffix(fileName, ".vm")
	filePath := strings.Split(fileNameWithoutExtension, "/")
	w.SetFileName(filePath[len(filePath) - 1])

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
		case "C_LABEL":
			labelCommand := fmt.Sprintf("%s$%s", p.CurrentFunction, arg1)
			err = w.WriteLabel(labelCommand)
		case "C_GOTO", "C_IF":
			branchCommand := fmt.Sprintf("%s$%s", p.CurrentFunction, arg1)
			if commandType == "C_GOTO" {
				err = w.WriteGoto(branchCommand)
			} else {
				err = w.WriteIf(branchCommand)
			}
		case "C_FUNCTION":
			err = w.WriteFunction(arg1, arg2)
		case "C_CALL":
			err = w.WriteCall(arg1, arg2)
		case "C_RETURN":
			err = w.WriteReturn()
		}
	}

	return err
}


func main() {
	args := os.Args
	if len(args) != 2 {
		fmt.Println("Usage: ./vm-translator <Filename>.vm")
		os.Exit(1)
	}
	inputPath := args[1]
	filePath := strings.Split(inputPath, "/")
	isFile := strings.Contains(filePath[len(filePath) - 1], ".")
	var err error
	var w *writer.Writer
	if isFile {
		if !unicode.IsUpper(rune(filePath[len(filePath) - 1][0])) || !strings.HasSuffix(inputPath, ".vm") {
			fmt.Println("Invalid filename - should be of the format: Filename.vm")
			os.Exit(1)
		}
		inputPathWithoutExtension, _ := strings.CutSuffix(inputPath, ".vm")
		w, err = writer.NewWriter(inputPathWithoutExtension + ".asm")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer w.Close()
		err = translate(inputPath, w)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		filePath = filePath[:len(filePath) - 1]
		w, err = writer.NewWriter(fmt.Sprintf("%s/%s.asm", inputPath, filePath[len(filePath) - 1]))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer w.Close()
		folder, err := os.Open(inputPath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fileNames, err := folder.Readdirnames(0)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		for _, fileName := range fileNames {
			if !unicode.IsUpper(rune(fileName[0])) || !strings.HasSuffix(fileName, ".vm") {
				continue
			}
			err = translate(inputPath + fileName, w)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}
}
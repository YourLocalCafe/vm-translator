package writer

import (
	"errors"
	"fmt"
	"os"
	"strings"
)


type Writer struct {
	file *os.File
	equalCounter int
	gtCounter int
	ltCounter int
	skipCounter int
}


const binaryCommands = `// Binary command
@SP
M=M-1
A=M
D=M
@SP
M=M-1
A=M
%s
@SP
M=M+1
`

const unaryCommands = `// Unary command
@SP
M=M-1
A=M
M=%sM
@SP
M=M+1
`

const compareCommands = `// Compare command
@SP
M=M-1
A=M
D=M
@SP
M=M-1
A=M
D=M-D
%s
@SP
A=M
M=0
@SP
M=M+1
%s
0;JMP
%s
@SP
A=M
M=-1
@SP
M=M+1
%s
`

const pushSegmentCommands = `// Push segment command
%s
D=M
%s
A=D+A
D=M
@SP
A=M
M=D
@SP
M=M+1
`

const popSegmentCommands = `// Pop segment command
%s
D=M
%s
D=D+A
@R13
M=D
@SP
M=M-1
A=M
D=M
@R13
A=M
M=D
`

const pushConstantCommands = `// Push constant command
%s
D=A
@SP
A=M
M=D
@SP
M=M+1
`

const pushPointerCommands = `// Push pointer command
%s
D=M
@SP
A=M
M=D
@SP
M=M+1
`

const popPointerCommands = `// Pop pointer command
@SP
M=M-1
A=M
D=M
%s
M=D
`

const pushTempCommands = `// Push temp command
@5
D=A
%s
A=D+A
D=M
@SP
A=M
M=D
@SP
M=M+1
`

const popTempCommands = `// Pop temp command
@5
D=A
%s
D=D+A
@R13
M=D
@SP
M=M-1
A=M
D=M
@R13
A=M
M=D
`

const pushStaticCommands = `// Push static command
%s
D=M
@SP
A=M
M=D
@SP
M=M+1
`

const popStaticCommands = `// Pop static command
@SP
M=M-1
A=M
D=M
%s
M=D
`

const endCommands = `@END
(END)
0;JMP
`

const gotoCommands = `// Goto command
@%s
0;JMP
`

const ifCommands = `// If-goto command
@SP
M=M-1
A=M
D=M
@%s
D;JNE
`


func NewWriter(filename string) (*Writer, error) {
	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	return &Writer{
		file: f,
	}, nil
}


func (w *Writer) WriteArithmetic(command string) error {
	switch command {
	case "add":
		return w.writeAdd()
	case "sub":
		return w.writeSub()
	case "neg":
		return w.writeNeg()
	case "and":
		return w.writeAnd()
	case "or":
		return w.writeOr()
	case "not":
		return w.writeNot()
	case "eq":
		return w.writeEq()
	case "gt":
		return w.writeGt()
	case "lt":
		return w.writeLt()
	default:
		errorMessage := fmt.Sprintf("An error occurred during code generation - Invalid command passed: %s", command)
		return errors.New(errorMessage)
	}
}


func (w *Writer) WritePushPop(commandType string, segment string, index int) error {
	switch commandType {
	case "C_PUSH":
		switch segment {
		case "local":
			return w.writePushSegment(segment, index)
		case "argument":
			return w.writePushSegment(segment, index)
		case "this":
			return w.writePushSegment(segment, index)
		case "that":
			return w.writePushSegment(segment, index)
		case "static":
			return w.writePushStatic(index)
		case "pointer":
			return w.writePushPointer(index)
		case "constant":
			return w.writePushConstant(index)
		case "temp":
			return w.writePushTemp(index)
		default:
			errorMessage := fmt.Sprintf("An error occurred during code generation - Invalid segment type: %s", segment)
			return errors.New(errorMessage)
		}
	case "C_POP":
		switch segment {
		case "local":
			return w.writePopSegment(segment, index)
		case "argument":
			return w.writePopSegment(segment, index)
		case "this":
			return w.writePopSegment(segment, index)
		case "that":
			return w.writePopSegment(segment, index)
		case "static":
			return w.writePopStatic(index)
		case "pointer":
			return w.writePopPointer(index)
		case "temp":
			return w.writePopTemp(index)
		default:
			errorMessage := fmt.Sprintf("An error occurred during code generation - Invalid segment type: %s", segment)
			return errors.New(errorMessage)
		}
	default:
		errorMessage := fmt.Sprintf("An error occurred during code generation - Invalid command type: %s", commandType)
		return errors.New(errorMessage)
	}
}


func (w *Writer) WriteIf(label string) error {
	_, err := fmt.Fprintf(w.file, ifCommands, label)
	return err
}


func (w *Writer) WriteGoto(label string) error {
	_, err := fmt.Fprintf(w.file, gotoCommands, label)
	return err
}


func (w *Writer) WriteLabel(label string) error {
	_, err := fmt.Fprintln(w.file, label)
	return err
}


func (w *Writer) writePopStatic(index int) error {
	filePath := strings.Split(w.file.Name(), "/")
	fileName, _, foundExtension := strings.Cut(filePath[len(filePath) - 1], ".")
	if !foundExtension {
		errorMessage := fmt.Sprintf("An error occurred during code generation - Invalid filename passed: %s", w.file.Name())
		return errors.New(errorMessage)
	}
	_, err := fmt.Fprintf(w.file, popStaticCommands, fmt.Sprintf("@%s.%d", fileName, index))
	return err
}


func (w *Writer) writePushStatic(index int) error {
	filePath := strings.Split(w.file.Name(), "/")
	fileName, _, foundExtension := strings.Cut(filePath[len(filePath) - 1], ".")
	if !foundExtension {
		errorMessage := fmt.Sprintf("An error occurred during code generation - Invalid filename passed: %s", w.file.Name())
		return errors.New(errorMessage)
	}
	_, err := fmt.Fprintf(w.file, pushStaticCommands, fmt.Sprintf("@%s.%d", fileName, index))
	return err
}


func (w *Writer) writePopTemp(index int) error {
	_, err := fmt.Fprintf(w.file, popTempCommands, fmt.Sprintf("@%d", index))
	return err
}


func (w *Writer) writePushTemp(index int) error {
	_, err := fmt.Fprintf(w.file, pushTempCommands, fmt.Sprintf("@%d", index))
	return err
}


func (w *Writer) writePopPointer(index int) error {
	var segment string
	switch index {
	case 0:
		segment = "THIS"
	case 1:
		segment = "THAT"
	default:
		errorMessage := fmt.Sprintf("An error occurred during code generation - Invalid index for pointer: %d", index)
		return errors.New(errorMessage)
	}
	_, err := fmt.Fprintf(w.file, popPointerCommands, fmt.Sprintf("@%s", segment))
	return err
}


func (w *Writer) writePushPointer(index int) error {
	var segment string
	switch index {
	case 0:
		segment = "THIS"
	case 1:
		segment = "THAT"
	default:
		errorMessage := fmt.Sprintf("An error occurred during code generation - Invalid index for pointer: %d", index)
		return errors.New(errorMessage)
	}
	_, err := fmt.Fprintf(w.file, pushPointerCommands, fmt.Sprintf("@%s", segment))
	return err
}


func (w *Writer) writePushConstant(value int) error {
	_, err := fmt.Fprintf(w.file, pushConstantCommands, fmt.Sprintf("@%d", value))
	return err
}


func (w *Writer) writePopSegment(segment string, index int) error {
	var segmentName string
	switch segment {
	case "local":
		segmentName = "LCL"
	case "argument":
		segmentName = "ARG"
	case "this":
		segmentName = "THIS"
	case "that":
		segmentName = "THAT"
	default:
		errorMessage := fmt.Sprintf("An error occurred during code generation - Invalid segment type passed: %s", segment)
		return errors.New(errorMessage)
	}
	_, err := fmt.Fprintf(w.file, popSegmentCommands, fmt.Sprintf("@%s", segmentName), fmt.Sprintf("@%d", index))
	return err
}


func (w *Writer) writePushSegment(segment string, index int) error {
	var segmentName string
	switch segment {
	case "local":
		segmentName = "LCL"
	case "argument":
		segmentName = "ARG"
	case "this":
		segmentName = "THIS"
	case "that":
		segmentName = "THAT"
	default:
		errorMessage := fmt.Sprintf("An error occurred during code generation - Invalid segment type passed: %s", segment)
		return errors.New(errorMessage)
	}
	_, err := fmt.Fprintf(w.file, pushSegmentCommands, fmt.Sprintf("@%s", segmentName), fmt.Sprintf("@%d", index))
	return err
}


func (w *Writer) writeAdd() error {
	_, err := fmt.Fprintf(w.file, binaryCommands, "M=D+M")
	return err
}


func (w *Writer) writeSub() error {
	_, err := fmt.Fprintf(w.file, binaryCommands, "M=M-D")
	return err
}


func (w *Writer) writeNeg() error {
	_, err := fmt.Fprintf(w.file, unaryCommands, "-")
	return err
}


func (w *Writer) writeAnd() error {
	_, err := fmt.Fprintf(w.file, binaryCommands, "M=D&M")
	return err
}


func (w *Writer) writeOr() error {
	_, err := fmt.Fprintf(w.file, binaryCommands, "M=D|M")
	return err
}


func (w *Writer) writeNot() error {
	_, err := fmt.Fprintf(w.file, unaryCommands, "!")
	return err
}


func (w *Writer) writeEq() error {
	finalCommand := fmt.Sprintf(compareCommands, fmt.Sprintf("@EQUAL%d\nD;JEQ", w.equalCounter), fmt.Sprintf("@SKIP%d", w.skipCounter), fmt.Sprintf("(EQUAL%d)", w.equalCounter), fmt.Sprintf("(SKIP%d)", w.skipCounter))
	_, err := fmt.Fprint(w.file, finalCommand)
	if err != nil {
		return err
	}
	w.equalCounter++
	w.skipCounter++
	return nil
}


func (w *Writer) writeGt() error {
	finalCommand := fmt.Sprintf(compareCommands, fmt.Sprintf("@GT%d\nD;JGT", w.gtCounter), fmt.Sprintf("@SKIP%d", w.skipCounter), fmt.Sprintf("(GT%d)", w.gtCounter), fmt.Sprintf("(SKIP%d)", w.skipCounter))
	_, err := fmt.Fprint(w.file, finalCommand)
	if err != nil {
		return err
	}
	w.gtCounter++
	w.skipCounter++
	return nil
}


func (w *Writer) writeLt() error {
	finalCommand := fmt.Sprintf(compareCommands, fmt.Sprintf("@LT%d\nD;JLT", w.ltCounter), fmt.Sprintf("@SKIP%d", w.skipCounter), fmt.Sprintf("(LT%d)", w.ltCounter), fmt.Sprintf("(SKIP%d)", w.skipCounter))
	_, err := fmt.Fprint(w.file, finalCommand)
	if err != nil {
		return err
	}
	w.ltCounter++
	w.skipCounter++
	return nil
}


func (w *Writer) Close() error {
	_, err := fmt.Fprint(w.file, endCommands)
	if err != nil {
		return err
	}
	err = w.file.Close()
	return err
}
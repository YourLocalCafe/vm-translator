package writer

import (
	"errors"
	"fmt"
	"os"
)


type Writer struct {
	file *os.File
	inputFileName string
	equalCounter int
	gtCounter int
	ltCounter int
	skipCounter int
	functionCounter int
	callCounter int
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
@%s.%s
0;JMP
`

const ifCommands = `// If-goto command
@SP
M=M-1
A=M
D=M
@%s.%s
D;JNE
`

const functionCommands = `// Function command
%s
@%d
D=A
@R13
M=D
@SKIP_LOCAL_LOOP%d
D;JEQ
(LOCAL_LOOP%d)
@SP
A=M
M=0
@SP
M=M+1
@R13
M=M-1
D=M
@LOCAL_LOOP%d
D;JGT
(SKIP_LOCAL_LOOP%d)
`

const callCommands = `// Call command
@%s$ret.%d
D=A
@SP
A=M
M=D
@SP
M=M+1
@LCL
D=M
@SP
A=M
M=D
@SP
M=M+1
@ARG
D=M
@SP
A=M
M=D
@SP
M=M+1
@THIS
D=M
@SP
A=M
M=D
@SP
M=M+1
@THAT
D=M
@SP
A=M
M=D
@SP
M=M+1
D=M
@5
D=D-A
@%d
D=D-A
@ARG
M=D
@SP
D=M
@LCL
M=D
@%s
0;JMP
(%s$ret.%d)
`

const returnCommands = `// Return command
@LCL
D=M
@R13
M=D
@5
D=D-A
A=D
D=M
@R14
M=D
@SP
M=M-1
A=M
D=M
@ARG
A=M
M=D
@ARG
D=M+1
@SP
M=D
@R13
D=M-1
A=D
D=M
@THAT
M=D
@R13
D=M
@2
D=D-A
A=D
D=M
@THIS
M=D
@R13
D=M
@3
D=D-A
A=D
D=M
@ARG
M=D
@R13
D=M
@4
D=D-A
A=D
D=M
@LCL
M=D
@R14
A=M
0;JMP
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


func (w *Writer) SetFileName(fileName string) {
	w.inputFileName = fileName
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


func (w *Writer) WriteReturn() error {
	_, err := fmt.Fprintf(w.file, returnCommands)
	return err
}


func (w *Writer) WriteCall(functionName string, nArgs int) error {
	functionLabel := fmt.Sprintf("%s.%s", w.inputFileName, functionName)
	_, err := fmt.Fprintf(w.file, callCommands, functionLabel, w.callCounter, nArgs, functionLabel, functionLabel, w.callCounter)
	w.callCounter++
	return err
}


func (w *Writer) WriteFunction(functionName string, nVars int) error {
	_, err := fmt.Fprintf(w.file, functionCommands, fmt.Sprintf("(%s.%s)", w.inputFileName, functionName), nVars, w.functionCounter, w.functionCounter, w.functionCounter, w.functionCounter)
	w.functionCounter++
	return err
}


func (w *Writer) WriteIf(label string) error {
	_, err := fmt.Fprintf(w.file, ifCommands, w.inputFileName, label)
	return err
}


func (w *Writer) WriteGoto(label string) error {
	_, err := fmt.Fprintf(w.file, gotoCommands, w.inputFileName, label)
	return err
}


func (w *Writer) WriteLabel(label string) error {
	_, err := fmt.Fprintln(w.file, fmt.Sprintf("(%s.%s)", w.inputFileName, label))
	return err
}


func (w *Writer) writePopStatic(index int) error {
	_, err := fmt.Fprintf(w.file, popStaticCommands, fmt.Sprintf("@%s.%d", w.inputFileName, index))
	return err
}


func (w *Writer) writePushStatic(index int) error {
	_, err := fmt.Fprintf(w.file, pushStaticCommands, fmt.Sprintf("@%s.%d", w.inputFileName, index))
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
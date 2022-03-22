package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Variable struct {
	Name  string
	Value uint8
}

type StringVariable struct {
	Name     string
	Value    string
	Constant bool
}

type FutureVariable struct {
	Name string
}

type Instruction struct {
	Opcode        string
	Args          string
	RegistersUsed []uint8
}

const (
	RegisterOutputIO = iota
	RegisterInputIO
	RegisterCharacterHolder
	RegisterMemoryBeginning
	RegisterGeneral
)

type Register struct {
	Name         uint8
	InUse        bool
	ToBeReleased bool
	Type         int
}

type Context struct {
	TemporaryRegistersInUse []Register
	SavedRegistersInUse     []Register
	Instructions            []*Instruction
	LoopCounter             int
}

type Program struct {
	code       []string
	variables  []Variable
	stringVars []StringVariable
	futureVars []FutureVariable
	contexts   []*Context
}

var ourProgram Program

func RemoveAtIndex(i int, f []FutureVariable) []FutureVariable {
	if i >= len(f) {
		return f
	}
	return append(f[:i], f[i+1:]...)
}

func parseLine(line string) (string, string) {
	// make line lowercase
	line = strings.ToLower(line)
	// if there are no spaces, then there is no args
	if strings.Index(line, " ") == -1 {
		return line, ""
	}
	// first word is the instruction
	instruction := line[:strings.Index(line, " ")]
	// rest is the arguments
	args := line[len(instruction)+1:]
	return instruction, args
}

// takes in an instruction and its arguments and returns the assembly code for it
func (c *Context) handleInstruction(instruction string, arguments string) error {
	switch instruction {
	case "print":
		return c.handlePrint(arguments)
	case "read":
		return c.handleRead(arguments)
	case "let":
		return handleLet(arguments)
	case "addi":
		return c.handleAddi(arguments)
	case "ret":
		c.AddInstruction(&Instruction{Opcode: "jr", Args: "$0", RegistersUsed: nil}, false)
		return nil
	default:
		// return error
		return fmt.Errorf("invalid instruction")
	}
}

func (c *Context) compileInstructions() []string {
	var assembly []string
	for _, instruction := range c.Instructions {
		assembly = append(assembly, fmt.Sprintf("%s %s", instruction.Opcode, instruction.Args))
	}
	return assembly
}

func main() {
	// first argument should be infile
	// second argument should be outfile (optional)
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go infile [outfile]")
		os.Exit(1)
	}

	infile := os.Args[1]
	outfile := os.Args[1] + ".mips" // compile to mips assembly code
	if len(os.Args) > 2 {
		outfile = os.Args[2]
	}

	mainContext := CreateMainContext()
	ourProgram.contexts = append(ourProgram.contexts, mainContext)

	// read in the file
	file, err := ioutil.ReadFile(infile)
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}

	// split the file into lines
	lines := strings.Split(string(file), "\n")

	// parse each line
	lineNum := 1
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		lineNum++
		// parse the line
		instruction, args := parseLine(line)

		// handle the instruction
		err := mainContext.handleInstruction(instruction, args)
		if err != nil {
			fmt.Println("Error handling instruction: ", err)
			fmt.Println("Line: ", lineNum)
			os.Exit(1)
		}
	}

	// compile the instructions
	for _, context := range ourProgram.contexts {
		context.compileInstructions()
	}

	// write the assembly code to the outfile
	err = ioutil.WriteFile(outfile, []byte(strings.Join(ourProgram.contexts[0].compileInstructions(), "\n")), 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		os.Exit(1)
	}
}

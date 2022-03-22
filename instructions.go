package main

import (
	"fmt"
	"strconv"
	"strings"
)

func handleLet(args string) error {
	// make sure there are two arguments
	argsArray := strings.Split(args, " ")
	if len(argsArray) != 2 {
		return fmt.Errorf("let: invalid number of arguments")
	}
	// first argument should be the variable name
	varName := argsArray[0]
	// second argument should be the value
	varValue := argsArray[1]

	// for debugging
	fmt.Printf("varName: %s\n", varName)
	fmt.Printf("varValue: %s\n", varValue)

	// if the value is future, then it will be assigned later
	if varValue == "future" {
		// add the variable to the future map
		ourProgram.futureVars = append(ourProgram.futureVars, FutureVariable{
			Name: varName,
		})
		return nil
	} else {
		// check if variable is a number
		if _, err := strconv.Atoi(varValue); err == nil {
			fmt.Println("is number")
			// if it is a number, then assign it to the variable
			i, err := strconv.Atoi(varValue)
			if err != nil {
				return fmt.Errorf("let: %s", err)
			}
			ourProgram.variables = append(ourProgram.variables, Variable{
				Name:  varName,
				Value: uint8(i),
			})
			return nil
		} else {
			// if it is not a number, then check for quotes
			if varValue[0] == '"' && varValue[len(varValue)-1] == '"' {
				// if it is a string, then assign it to the variable
				ourProgram.stringVars = append(ourProgram.stringVars, StringVariable{
					Name:  varName,
					Value: varValue[1 : len(varValue)-1],
				})
				return nil
			} else {
				// throw error
				return fmt.Errorf("let: invalid value: %s", varValue)
			}
		}
	}
}

func (c *Context) handleAddi(args string) error {
	// args should be var1, var2, number
	// or var1, 0, number
	argsArray := strings.Split(args, " ")
	if len(argsArray) != 3 {
		// if two args, then it is just var1 + number
		if len(argsArray) == 2 {
			varname := argsArray[0]
			number, err := strconv.Atoi(argsArray[1])
			if err != nil {
				return fmt.Errorf("addi: %s", err)
			}
			for i := 0; i < len(ourProgram.variables); i++ {
				if ourProgram.variables[i].Name == varname {
					ourProgram.variables[i].Value += uint8(number)
					return nil
				}
			}
		} else {
			return fmt.Errorf("addi: invalid number of arguments")
		}
	}
	// first argument should be a variable name
	varAname := argsArray[0]
	// second argument should be a variable name
	varBname := argsArray[1]
	// third argument should be a number
	varValue, err := strconv.Atoi(argsArray[2])
	if err != nil {
		return fmt.Errorf("addi: invalid number")
	}
	tmpResult := 0
	// check if variable B is 0
	if varBname == "0" {
		// todo
	}
	// check if variable B is a number (it cannot be a future variable)
	for _, variable := range ourProgram.variables {
		if variable.Name == varBname {
			// find variable A
			for _, variableA := range ourProgram.variables {
				if variableA.Name == varAname {
					// add the two variables
					tmpResult = int(variable.Value) + varValue
					// check if the result is a number
					if tmpResult > 255 {
						return fmt.Errorf("addi: result is too large")
					}
					// assign the result to the variable
					variableA.Value = uint8(tmpResult)
					return nil
				}
			}
			// if we're still here, variable A might be a future variable
			// if so, remove it from futures and add it to variables
			for i, futureVariable := range ourProgram.futureVars {
				if futureVariable.Name == varAname {
					ourProgram.variables = append(ourProgram.variables, Variable{
						Name:  varAname,
						Value: uint8(variable.Value + uint8(varValue)),
					})
					ourProgram.futureVars = append(ourProgram.futureVars[:i], ourProgram.futureVars[i+1:]...)
					return nil
				}
			}
			// variable A doesn't exist
			return fmt.Errorf("addi: variable does not exist: %s", varAname)
		}
	}
	return fmt.Errorf("addi: variable does not exist: %s", varBname)
}

func (c *Context) handleRead(args string) error {
	// args should be a variable to put the result in
	varName := args
	varPlace := 0
	var register uint8

	// check if variable exists
	for i, variable := range ourProgram.stringVars {
		if variable.Name == varName {
			// if not constant, use its register
			if variable.Constant == false {
				varPlace = i
				// remove first two characters to get register
				tmp, err := strconv.Atoi(variable.Value[2:])
				if err != nil {
					return fmt.Errorf("read: %s", err)
				}
				register = uint8(tmp)
				break
			} else {
				varPlace = i
				ourProgram.stringVars[i].Value = "$s" + strconv.Itoa(int(c.FindUnusedSavedRegister()))
			}
		}
	}
	// check future variables
	for i, futureVariable := range ourProgram.futureVars {
		if futureVariable.Name == varName {
			register = c.FindUnusedSavedRegister()
			// move to string variables
			ourProgram.stringVars = append(ourProgram.stringVars, StringVariable{
				Name:     varName,
				Value:    "$s" + strconv.Itoa(int(register)),
				Constant: false,
			})
			// remove future variable
			ourProgram.futureVars = append(ourProgram.futureVars[:i], ourProgram.futureVars[i+1:]...)
			varPlace = len(ourProgram.stringVars) - 1
		}
	}

	// to get string, read from mem address 0x3000
	// then, loop and store in register + loop counter until we read -1

	ioAddressRegister, preExisting := c.FindUnusedTemporaryRegister(RegisterInputIO)
	characterRegister, _ := c.FindUnusedTemporaryRegister(RegisterCharacterHolder)
	counterRegister, _ := c.FindUnusedTemporaryRegister(RegisterGeneral)

	if !preExisting {
		c.AddInstruction(&Instruction{
			Opcode: "lui",
			Args:   fmt.Sprintf("$t%d, %s", ioAddressRegister, "0x3000"),
			RegistersUsed: []uint8{
				ioAddressRegister,
			},
		}, false)
	}

	memoryBeginningRegister, _ := c.FindUnusedTemporaryRegister(RegisterGeneral)
	c.AddInstruction(&Instruction{
		Opcode: "lui",
		Args:   fmt.Sprintf("$t%d, %s", memoryBeginningRegister, "0x2000"),
		RegistersUsed: []uint8{
			ioAddressRegister,
		},
	}, false)

	// set counter to 0

	c.AddInstruction(&Instruction{
		Opcode: "addi",
		Args:   fmt.Sprintf("$t%d, $0, 0", counterRegister),
		RegistersUsed: []uint8{
			counterRegister,
		},
	}, false)

	// new loop

	loopName := fmt.Sprintf("loop%d", c.LoopCounter)
	c.LoopCounter++
	c.AddInstruction(&Instruction{
		Opcode:        loopName + ":",
		Args:          "",
		RegistersUsed: []uint8{},
	}, false)

	// load character from memory

	c.AddInstruction(&Instruction{
		Opcode: "lw",
		Args:   fmt.Sprintf("$t%d, 0($t%d)", characterRegister, ioAddressRegister),
		RegistersUsed: []uint8{
			ioAddressRegister,
			characterRegister,
		},
	}, false)

	// convert from 32 bit to 8 bit

	c.AddInstruction(&Instruction{
		Opcode: "sll",
		Args:   fmt.Sprintf("$t%d, $t%d, 0", characterRegister, characterRegister),
		RegistersUsed: []uint8{
			characterRegister,
		},
	}, false)

	// store character in data memory

	c.AddInstruction(&Instruction{
		Opcode: "sb",
		Args:   fmt.Sprintf("$t%d, 0($t%d)", characterRegister, memoryBeginningRegister),
		RegistersUsed: []uint8{
			counterRegister,
			characterRegister,
		},
	}, false)

	// test if character is less than 31
	c.AddInstruction(&Instruction{
		Opcode: "slti",
		Args:   fmt.Sprintf("$t%d, $t%d, 31", characterRegister, characterRegister),
		RegistersUsed: []uint8{
			characterRegister,
		},
	}, false)

	c.AddInstruction(&Instruction{
		Opcode: "bgtz",
		Args:   fmt.Sprintf("$t%d, %s", characterRegister, loopName+"-end"),
		RegistersUsed: []uint8{
			characterRegister,
		},
	}, false)

	// add one to the counter as well as the address

	c.AddInstruction(&Instruction{
		Opcode: "addi",
		Args:   fmt.Sprintf("$t%d, $t%d, 1", counterRegister, counterRegister),
		RegistersUsed: []uint8{
			counterRegister,
		},
	}, false)

	c.AddInstruction(&Instruction{
		Opcode: "addi",
		Args:   fmt.Sprintf("$t%d, $t%d, 1", memoryBeginningRegister, memoryBeginningRegister),
		RegistersUsed: []uint8{
			ioAddressRegister,
		},
	}, false)

	// j loopName
	c.AddInstruction(&Instruction{
		Opcode: "j",
		Args:   loopName,
		RegistersUsed: []uint8{
			characterRegister,
		},
	}, false)

	// nop

	c.AddInstruction(&Instruction{
		Opcode:        "nop",
		Args:          "",
		RegistersUsed: []uint8{},
	}, false)

	// (loopName)-end
	c.AddInstruction(&Instruction{
		Opcode:        loopName + "-end:",
		Args:          "",
		RegistersUsed: []uint8{},
	}, false)

	// set (register) to (counterRegister)

	c.AddInstruction(&Instruction{
		Opcode: "addi",
		Args:   fmt.Sprintf("$s%d, $t%d, 0", register, counterRegister),
		RegistersUsed: []uint8{
			counterRegister,
		},
	}, false)
	// set (stringVar) to (register)
	ourProgram.stringVars[varPlace].Value = "$s" + strconv.Itoa(int(register))
	fmt.Println(ourProgram.stringVars[varPlace])
	// free the registers
	c.ReleaseTemporaryRegister(ioAddressRegister)
	c.ReleaseTemporaryRegister(characterRegister)
	c.ReleaseTemporaryRegister(counterRegister)
	c.ReleaseTemporaryRegister(memoryBeginningRegister)
	return nil
}

func (c *Context) handlePrint(s string) error {
	stringA := ""
	constVar := true
	// check if followed by a quote
	if s[0] == '"' && s[len(s)-1] == '"' {
		// remove quotes and add to string
		stringA = s[1 : len(s)-1]
	} else if s[0] >= '0' && s[0] <= '9' {
		// are all digits?
		if strings.Index(s, " ") == -1 {
			// yes, convert to int and add to string
			stringA = fmt.Sprintf("%d", s)
		} else {
			// no, return error
			return fmt.Errorf("print: invalid argument")
		}
	} else if len(s) > 0 {
		// check if variable exists
		for _, variable := range ourProgram.variables {
			if variable.Name == s {
				// yes, add to string
				stringA = fmt.Sprintf("%d", variable.Value)
			}
		}
		// check if string variable exists
		for _, variable := range ourProgram.stringVars {
			if variable.Name == s {
				// yes, add to string
				stringA = variable.Value
				constVar = variable.Constant
			}
		}
	} else {
		// no, return error
		return fmt.Errorf("print: invalid argument")
	}

	// now time to convert to assembly
	// for each character, get the ascii value
	var asciiCodes []int
	for _, char := range stringA {
		asciiCodes = append(asciiCodes, int(char))
	}
	/*
		we need to generate
		addi (unused register), $0, (ascii code)
		sw (previous register), 0(mem address of output)
	*/
	// addi (unused register), $0, (ascii code)
	ioReg, preExisting := c.FindUnusedTemporaryRegister(RegisterOutputIO)

	if !preExisting {
		// setup the register
		c.AddInstruction(&Instruction{
			Opcode:        "addi",
			Args:          fmt.Sprintf("$t%d, $0, 0x30", ioReg),
			RegistersUsed: []uint8{ioReg},
		}, false)
		c.AddInstruction(&Instruction{
			Opcode:        "sll",
			Args:          fmt.Sprintf("$t%d, $t%d, 24", ioReg, ioReg),
			RegistersUsed: []uint8{ioReg},
		}, false)
		c.AddInstruction(&Instruction{
			Opcode:        "addi",
			Args:          fmt.Sprintf("$t%d, $t%d, 4", ioReg, ioReg),
			RegistersUsed: []uint8{ioReg},
		}, false)
	}

	tmpCharReg, _ := c.FindUnusedTemporaryRegister(RegisterCharacterHolder)
	// if we're using a string variable and it is not constant, we need to load it (value will be the register with its length)
	if !constVar {
		fmt.Println("not constant")
		memoryBeginningRegister, _ := c.FindUnusedTemporaryRegister(RegisterGeneral)
		counterRegister, _ := c.FindUnusedTemporaryRegister(RegisterGeneral)

		// get a pointer to the beginning of the memory area

		c.AddInstruction(&Instruction{
			Opcode: "lui",
			Args:   fmt.Sprintf("$t%d, %s", memoryBeginningRegister, "0x2000"),
			RegistersUsed: []uint8{
				memoryBeginningRegister,
			},
		}, false)

		// set the counterRegister to value of stringA ($s0)

		c.AddInstruction(&Instruction{
			Opcode:        "add",
			Args:          fmt.Sprintf("$t%d, $0, %s", counterRegister, stringA),
			RegistersUsed: []uint8{counterRegister},
		}, false)

		// new loop

		loopName := fmt.Sprintf("loop%d", c.LoopCounter)
		c.LoopCounter++
		c.AddInstruction(&Instruction{
			Opcode:        loopName + ":",
			Args:          "",
			RegistersUsed: []uint8{},
		}, false)

		// branch if counter is 0

		c.AddInstruction(&Instruction{
			Opcode:        "beq",
			Args:          fmt.Sprintf("$t%d, $0, %s", counterRegister, loopName+"-end"),
			RegistersUsed: []uint8{counterRegister},
		}, false)

		// load byte from memory

		c.AddInstruction(&Instruction{
			Opcode:        "lb",
			Args:          fmt.Sprintf("$t%d, 0($t%d)", tmpCharReg, memoryBeginningRegister),
			RegistersUsed: []uint8{memoryBeginningRegister, tmpCharReg},
		}, false)

		// increment memory address

		c.AddInstruction(&Instruction{
			Opcode:        "addi",
			Args:          fmt.Sprintf("$t%d, $t%d, 1", memoryBeginningRegister, memoryBeginningRegister),
			RegistersUsed: []uint8{memoryBeginningRegister},
		}, false)

		// subtract 1 from counter

		veryTemporaryRegister, _ := c.FindUnusedTemporaryRegister(RegisterGeneral)

		c.AddInstruction(&Instruction{
			Opcode:        "addi",
			Args:          fmt.Sprintf("$t%d, $0, 1", veryTemporaryRegister),
			RegistersUsed: []uint8{veryTemporaryRegister},
		}, false)

		c.AddInstruction(&Instruction{
			Opcode:        "sub",
			Args:          fmt.Sprintf("$t%d, $t%d, $t%d", counterRegister, counterRegister, veryTemporaryRegister),
			RegistersUsed: []uint8{counterRegister, veryTemporaryRegister},
		}, false)

		c.ReleaseTemporaryRegister(veryTemporaryRegister)

		// print byte

		c.AddInstruction(&Instruction{
			Opcode:        "sw",
			Args:          fmt.Sprintf("$t%d, 0($t%d)", tmpCharReg, ioReg),
			RegistersUsed: []uint8{ioReg, tmpCharReg},
		}, false)

		// jump to loop

		c.AddInstruction(&Instruction{
			Opcode:        "j",
			Args:          loopName,
			RegistersUsed: []uint8{},
		}, false)

		// nop

		c.AddInstruction(&Instruction{
			Opcode:        "nop",
			Args:          "",
			RegistersUsed: []uint8{},
		}, false)

		c.AddInstruction(&Instruction{
			Opcode:        loopName + "-end:",
			Args:          "",
			RegistersUsed: []uint8{},
		}, false)
		// free the register
		c.ReleaseTemporaryRegister(counterRegister)
		c.ReleaseTemporaryRegister(ioReg)
		c.ReleaseTemporaryRegister(tmpCharReg)

		return nil
	} else {

		// for each ascii code, generate assembly
		for _, asciiCode := range asciiCodes {
			c.AddInstruction(&Instruction{
				Opcode:        "addi",
				Args:          fmt.Sprintf("$t%d, $0, %d", tmpCharReg, asciiCode),
				RegistersUsed: []uint8{tmpCharReg},
			}, false)
			c.AddInstruction(&Instruction{
				Opcode:        "sw",
				Args:          fmt.Sprintf("$t%d, 0($t%d)", tmpCharReg, ioReg),
				RegistersUsed: []uint8{ioReg, tmpCharReg},
			}, false)
		}

		// print a newline
		c.AddInstruction(&Instruction{
			Opcode:        "addi",
			Args:          fmt.Sprintf("$t%d, $0, 10", tmpCharReg),
			RegistersUsed: []uint8{ioReg},
		}, false)
		c.AddInstruction(&Instruction{
			Opcode:        "sw",
			Args:          fmt.Sprintf("$t%d, 0($t%d)", tmpCharReg, ioReg),
			RegistersUsed: []uint8{ioReg, ioReg},
		}, false)

		// free the registers
		c.ReleaseTemporaryRegister(ioReg)
		c.ReleaseTemporaryRegister(tmpCharReg)

		return nil
	}
}

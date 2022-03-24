package main

import (
	"fmt"
	"strconv"
	"strings"
)

// goto if
func (c *Context) handleGif(args string) error {
	// should have three arguments, first being a variable,
	// second being either a string literal, number literal, or variable, and third being a label
	argsArray := strings.Split(args, " ")
	if len(argsArray) != 3 {
		return fmt.Errorf("invalid number of arguments for gif")
	}
	// make sure first argument is a variable or string variable
	for _, varA := range ourProgram.variables {
		if argsArray[0] == varA.Name {
			// is the second argument a string or number literal?
			if argsArray[1][0] == '"' && argsArray[1][len(argsArray[1])-1] == '"' {
				// throw error
				return fmt.Errorf("gif: second argument cannot be a string literal if first arg is a number")
			} else if varB, ok := strconv.Atoi(argsArray[1]); ok == nil {
				// is varA constant?
				if varA.Constant {
					if varA.Value == uint8(varB) {
						// add jump instruction
						c.AddInstruction(&Instruction{
							Opcode:        "j",
							Args:          argsArray[2],
							RegistersUsed: nil,
						}, false)
						// add nop
						c.AddInstruction(&Instruction{
							Opcode:        "nop",
							Args:          "",
							RegistersUsed: nil,
						}, false)
						return nil
					}
				} else {
					// varA should be an $s register
					// load varB into a temporary register
					tmpReg, _ := c.FindUnusedTemporaryRegister(RegisterGeneral)
					c.AddInstruction(&Instruction{
						Opcode:        "li",
						Args:          fmt.Sprintf("$t%d, %d", tmpReg, varB),
						RegistersUsed: []uint8{tmpReg},
					}, false)
					// beq $s, $t, label
					c.AddInstruction(&Instruction{
						Opcode:        "beq",
						Args:          fmt.Sprintf("$s%d, $t%d, %s", varA.Value, tmpReg, argsArray[2]),
						RegistersUsed: []uint8{varA.Value, tmpReg},
					}, false)
					// free temporary register
					c.ReleaseTemporaryRegister(tmpReg)
				}
			} else {
				// is second argument a variable?
				for _, varB := range ourProgram.variables {
					if argsArray[1] == varB.Name {
						// is varB constant?
						if varB.Constant {
							// is varA constant?
							if varA.Constant {
								if varA.Value == varB.Value {
									// add jump instruction
									c.AddInstruction(&Instruction{
										Opcode:        "j",
										Args:          argsArray[2],
										RegistersUsed: nil,
									}, false)
									// add nop
									c.AddInstruction(&Instruction{
										Opcode:        "nop",
										Args:          "",
										RegistersUsed: nil,
									}, false)
									return nil
								}
							} else {
								// varA should be an $s register
								// load varB into a temporary register
								tmpReg, _ := c.FindUnusedTemporaryRegister(RegisterGeneral)
								c.AddInstruction(&Instruction{
									Opcode:        "li",
									Args:          fmt.Sprintf("$t%d, %d", tmpReg, varB.Value),
									RegistersUsed: []uint8{tmpReg},
								}, false)
								// beq $s, $t, label
								c.AddInstruction(&Instruction{
									Opcode:        "beq",
									Args:          fmt.Sprintf("$s%d, $t%d, %s", varA.Value, tmpReg, argsArray[2]),
									RegistersUsed: []uint8{varA.Value, tmpReg},
								}, false)
								c.ReleaseTemporaryRegister(tmpReg)
							}

						} else {
							// is varA constant?
							if varA.Constant {
								// load varA into a temporary register
								tmpReg, _ := c.FindUnusedTemporaryRegister(RegisterGeneral)
								c.AddInstruction(&Instruction{
									Opcode:        "li",
									Args:          fmt.Sprintf("$t%d, %d", tmpReg, varA.Value),
									RegistersUsed: []uint8{tmpReg},
								}, false)
								// beq $s, $t, label
								c.AddInstruction(&Instruction{
									Opcode:        "beq",
									Args:          fmt.Sprintf("$s%d, $t%d, %s", varB.Value, tmpReg, argsArray[2]),
									RegistersUsed: []uint8{varB.Value, tmpReg},
								}, false)
								c.ReleaseTemporaryRegister(tmpReg)
							} else {
								// varA should be an $s register
								// varB should be an $s register
								// beq $s, $s, label
								c.AddInstruction(&Instruction{
									Opcode:        "beq",
									Args:          fmt.Sprintf("$s%d, $s%d, %s", varA.Value, varB.Value, argsArray[2]),
									RegistersUsed: []uint8{varA.Value, varB.Value},
								}, false)
							}
						}
					}
				}
			}
			return nil
		}
	}

	// if still here, look for a string variable
	for _, varA := range ourProgram.stringVars {
		if varA.Name == argsArray[0] {
			// is varB a string or number literal?
			needToWriteString := false
			if argsArray[1][0] == '"' && argsArray[1][len(argsArray[1])-1] == '"' {
				// if varA is constant, compare
				if varA.Constant {
					if varA.Value == argsArray[1][1:len(argsArray)-1] {
						c.AddInstruction(&Instruction{
							Opcode:        "j",
							Args:          argsArray[2],
							RegistersUsed: nil,
						}, false)
						// add nop
						c.AddInstruction(&Instruction{
							Opcode:        "nop",
							Args:          "",
							RegistersUsed: nil,
						}, false)
					}
				} else {
					needToWriteString = true

				}
			} else if _, ok := strconv.Atoi(argsArray[1]); ok == nil {
				return fmt.Errorf("gif: second argument cannot be a number literal if first arg is a string")
			} else {
			}
			// setup memory beginning register
			memoryBeginningRegister, _ := c.FindUnusedTemporaryRegister(RegisterMemoryBeginning)
			c.AddInstruction(&Instruction{
				Opcode: "lui",
				Args:   fmt.Sprintf("$t%d, %s", memoryBeginningRegister, "0x2000"),
				RegistersUsed: []uint8{
					memoryBeginningRegister,
				},
			}, false)
			if needToWriteString {
				varB := argsArray[1][1 : len(argsArray)-1]
				// add varA to $t
				c.AddInstruction(&Instruction{
					Opcode: "add",
					Args:   fmt.Sprintf("$t%d, $t%d, %s", memoryBeginningRegister, memoryBeginningRegister, varA.Value),
					RegistersUsed: []uint8{
						memoryBeginningRegister,
					},
				}, false)

				// for length of varB, add the following
				// li $t0, <char at i>
				// sb, $t0, i(membeginning)
				characterHolder, _ := c.FindUnusedTemporaryRegister(RegisterCharacterHolder)
				for i, char := range varB {
					c.AddInstruction(&Instruction{
						Opcode:        "li",
						Args:          fmt.Sprintf("$t%d, %d", characterHolder, char),
						RegistersUsed: []uint8{characterHolder},
					}, false)
					c.AddInstruction(&Instruction{
						Opcode:        "sb",
						Args:          fmt.Sprintf("$t%d, %d($t%d)", characterHolder, i, memoryBeginningRegister),
						RegistersUsed: []uint8{characterHolder, memoryBeginningRegister},
					}, false)
				}
				c.ReleaseTemporaryRegister(characterHolder)
			}

			// create second memory register
			memoryRegisterA, _ := c.FindUnusedTemporaryRegister(RegisterGeneral)
			c.AddInstruction(&Instruction{
				Opcode: "lui",
				Args:   fmt.Sprintf("$t%d, %s", memoryBeginningRegister, "0x2000"),
				RegistersUsed: []uint8{
					memoryBeginningRegister,
				},
			}, false)

			// loop iteration register
			counterRegister, _ := c.FindUnusedTemporaryRegister(RegisterGeneral)
			c.AddInstruction(&Instruction{
				Opcode:        "li",
				Args:          fmt.Sprintf("$t%d, 0", counterRegister),
				RegistersUsed: []uint8{counterRegister},
			}, false)

			// create a new loop to compare the strings
			loopName := fmt.Sprintf("loop%d", c.LoopCounter)
			c.LoopCounter++
			c.AddInstruction(&Instruction{
				Opcode:        loopName + ":",
				Args:          "",
				RegistersUsed: nil,
			}, false)
			// two character holders
			characterA, _ := c.FindUnusedTemporaryRegister(RegisterGeneral)
			characterB, _ := c.FindUnusedTemporaryRegister(RegisterGeneral)
			// load characterA & characterB
			c.AddInstruction(&Instruction{
				Opcode:        "lb",
				Args:          fmt.Sprintf("$t%d, 0($t%d)", characterA, memoryRegisterA),
				RegistersUsed: []uint8{characterA, memoryRegisterA},
			}, false)
			c.AddInstruction(&Instruction{
				Opcode:        "lb",
				Args:          fmt.Sprintf("$t%d, 0($t%d)", characterB, memoryBeginningRegister),
				RegistersUsed: []uint8{characterB, memoryBeginningRegister},
			}, false)

			// compare
			c.AddInstruction(&Instruction{
				Opcode:        "bne",
				Args:          fmt.Sprintf("$t%d, $t%d, %s", characterA, characterB, loopName+"-fail"),
				RegistersUsed: []uint8{characterB, characterA},
			}, true)

			// add one to each memory register
			c.AddInstruction(&Instruction{
				Opcode:        "addi",
				Args:          fmt.Sprintf("$t%d, $t%d, 1", memoryBeginningRegister, memoryBeginningRegister),
				RegistersUsed: []uint8{memoryBeginningRegister},
			}, false)
			c.AddInstruction(&Instruction{
				Opcode:        "addi",
				Args:          fmt.Sprintf("$t%d, $t%d, 1", memoryRegisterA, memoryRegisterA),
				RegistersUsed: []uint8{memoryRegisterA},
			}, false)

			// bge if counter is equal to string length
			c.AddInstruction(&Instruction{
				Opcode:        "bge",
				Args:          fmt.Sprintf("%s, $t%d, %s", varA.Value, counterRegister, loopName+"-end"),
				RegistersUsed: nil,
			}, false)

			// add one to counter
			c.AddInstruction(&Instruction{
				Opcode:        "addi",
				Args:          fmt.Sprintf("$t%d, $t%d, 1", counterRegister, counterRegister),
				RegistersUsed: []uint8{counterRegister},
			}, false)

			// jump to beginning of loop
			c.AddInstruction(&Instruction{
				Opcode:        "j",
				Args:          loopName,
				RegistersUsed: nil,
			}, false)
			// add nop
			c.AddInstruction(&Instruction{
				Opcode:        "nop",
				Args:          "",
				RegistersUsed: nil,
			}, false)

			// other sections

			// success
			c.AddInstruction(&Instruction{
				Opcode:        loopName + "-end:",
				Args:          "",
				RegistersUsed: nil,
			}, false)

			// jump to label
			c.AddInstruction(&Instruction{
				Opcode:        "j",
				Args:          argsArray[2],
				RegistersUsed: nil,
			}, false)

			// add nop
			c.AddInstruction(&Instruction{
				Opcode:        "nop",
				Args:          "",
				RegistersUsed: nil,
			}, false)

			// failure, continue on
			c.AddInstruction(&Instruction{
				Opcode:        loopName + "-fail:",
				Args:          "",
				RegistersUsed: nil,
			}, false)
		}
	}
	return nil
}

func (c *Context) handleLabel(label string) error {
	if c.DoesLabelExist(label) {
		return fmt.Errorf("label %s already exists", label)
	}
	c.ExistingLabels = append(c.ExistingLabels, label)
	// remove first character from label
	label = label[1:]
	// is length > 0
	if len(label) <= 0 {
		return fmt.Errorf("label %s is empty", label)
	}
	c.AddInstruction(&Instruction{
		Opcode:        label + ":",
		Args:          "",
		RegistersUsed: nil,
	}, false)
	return nil
}

func (c *Context) handleGoto(args string) error {
	if !c.DoesLabelExist(args) {
		return fmt.Errorf("label %s does not exist", args)
	}
	c.AddInstruction(&Instruction{
		Opcode:        "j",
		Args:          args,
		RegistersUsed: nil,
	}, false)
	// add nop
	c.AddInstruction(&Instruction{
		Opcode:        "nop",
		Args:          "",
		RegistersUsed: nil,
	}, false)
	return nil
}

func (c *Context) handleLen(args string) error {
	argsArray := strings.Split(args, " ")
	var value interface{}
	if len(argsArray) != 2 {
		return fmt.Errorf("len expects 2 arguments, got %d", len(argsArray))
	}
	// second argument is either a string or a stringVariable
	if argsArray[1][0] == '"' && argsArray[1][len(argsArray[1])-1] == '"' {
		// string literal
		// remove quotes and count the length
		value = len(argsArray[1][1 : len(argsArray[1])-1])
	} else {
		// possibly a string variable, check
		for _, v := range ourProgram.stringVars {
			if v.Name == argsArray[1] {
				// is constant?
				if v.Constant {
					value = len(v.Value)
				} else {
					// is variable, should be a register holding the length
					value = v.Value
				}
			}
		}
	}

	// first arg should be a variable
	for i, v := range ourProgram.variables {
		if v.Name == argsArray[0] {
			// is variable a constant?
			if v.Constant {
				// if value isn't an int, change to not constant
				if _, ok := value.(int); !ok {
					v.Constant = false
				}
			}
			// if value is an int, set it; otherwise, remove first two characters from value and convert to int
			if valueInt, ok := value.(uint8); ok {
				ourProgram.variables[i].Value = valueInt
				return nil
			} else if valueString, ok := value.(string); ok {
				tmp, err := strconv.Atoi(valueString[2:])
				if err != nil {
					return err
				}
				ourProgram.variables[i].Value = uint8(tmp)
				return nil
			} else {
				return fmt.Errorf("len expects a string or a string variable, got %T", value)
			}
		}
	}
	return nil
}

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
					// if variable isn't constant, then its value will be an $s register
					if !ourProgram.variables[i].Constant {
						c.AddInstruction(&Instruction{
							Opcode:        "addi",
							Args:          fmt.Sprintf("$s%d, $s%d, %d", i, i, number),
							RegistersUsed: []uint8{uint8(i)},
						}, false)
					} else {
						ourProgram.variables[i].Value += uint8(number)
					}
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
	for i, variable := range ourProgram.variables {
		if variable.Name == varBname {
			// find variable A
			for j, variableA := range ourProgram.variables {
				if variableA.Name == varAname {
					// if variable A is not constant, then its value will be an $s register
					if !variableA.Constant {
						// is variable B constant?
						if variable.Constant {
							// temporary register to hold the value of variable B
							tmpReg, _ := c.FindUnusedTemporaryRegister(RegisterGeneral)
							c.AddInstruction(&Instruction{
								Opcode:        "li",
								Args:          fmt.Sprintf("$t%d, %d", tmpReg, variable.Value),
								RegistersUsed: []uint8{tmpReg},
							}, false)
							// addi $s0, $t0, immediate
							c.AddInstruction(&Instruction{
								Opcode:        "addi",
								Args:          fmt.Sprintf("$s%d, $t%d, %d", ourProgram.variables[j].Value, tmpReg, varValue),
								RegistersUsed: []uint8{uint8(j), tmpReg},
							}, false)
							// free the temporary register
							c.ReleaseTemporaryRegister(tmpReg)
						} else {
							c.AddInstruction(&Instruction{
								Opcode:        "addi",
								Args:          fmt.Sprintf("$s%d, $s%d, %d", ourProgram.variables[i].Value, ourProgram.variables[j].Value, varValue),
								RegistersUsed: []uint8{uint8(j), uint8(i)},
							}, false)
						}
					} else {
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
			}
			// if we're still here, variable A might be a future variable
			// if so, remove it from futures and add it to variables
			for i, futureVariable := range ourProgram.futureVars {
				if futureVariable.Name == varAname {
					ourProgram.variables = append(ourProgram.variables, Variable{
						Name:     varAname,
						Value:    variable.Value,
						Constant: variable.Constant,
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

	memoryBeginningRegister, _ := c.FindUnusedTemporaryRegister(RegisterMemoryBeginning)
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

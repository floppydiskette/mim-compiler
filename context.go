package main

import "fmt"

func CreateMainContext() *Context {
	// make map of temporary registers in use
	var tmpRegisters []Register
	// set each to false
	var i uint8
	for i = 0; i <= 9; i++ {
		tmpRegisters = append(tmpRegisters, Register{i, false, false, -1})
	}
	// make map of saved registers
	var savedRegisters []Register
	// set each to false
	for i = 0; i <= 9; i++ {
		savedRegisters = append(savedRegisters, Register{i, false, false, -1})
	}
	return &Context{
		TemporaryRegistersInUse: tmpRegisters,
		SavedRegistersInUse:     savedRegisters,
		Instructions:            []*Instruction{},
	}
}

// if returns true, then we already have a good register
func (c *Context) FindUnusedTemporaryRegister(typeOf int) (uint8, bool) {
	// if type is RegisterOutputIO, find either used or unused register
	for i := 0; i <= 9; i++ {
		if c.TemporaryRegistersInUse[i].Type == typeOf && c.TemporaryRegistersInUse[i].Type != RegisterGeneral {
			fmt.Println("Found good IO register: ", i)
			c.TemporaryRegistersInUse[i].ToBeReleased = false
			c.TemporaryRegistersInUse[i].InUse = true
			c.TemporaryRegistersInUse[i].Type = typeOf
			return uint8(i), true
		}
	}
	for i := 0; i <= 9; i++ {
		if c.TemporaryRegistersInUse[i].InUse == false {
			c.TemporaryRegistersInUse[i].InUse = true
			c.TemporaryRegistersInUse[i].Type = typeOf
			return uint8(i), false
		}
	}
	for i := 0; i <= 9; i++ {
		if c.TemporaryRegistersInUse[i].ToBeReleased == true {
			c.TemporaryRegistersInUse[i].ToBeReleased = false
			c.TemporaryRegistersInUse[i].InUse = true
			c.TemporaryRegistersInUse[i].Type = typeOf
			return uint8(i), false
		}
	}
	return 200, false
}

func (c *Context) FindUnusedSavedRegister() uint8 {
	for i := 0; i <= 9; i++ {
		if c.SavedRegistersInUse[i].InUse == false {
			c.SavedRegistersInUse[i].InUse = true
			return uint8(i)
		}
	}
	for i := 0; i <= 9; i++ {
		if c.SavedRegistersInUse[i].ToBeReleased == true {
			c.SavedRegistersInUse[i].ToBeReleased = false
			c.SavedRegistersInUse[i].InUse = true
			return uint8(i)
		}
	}
	return 200
}

func (c *Context) ReleaseSavedRegister(reg uint8) {
	c.SavedRegistersInUse[reg].InUse = false
	c.SavedRegistersInUse[reg].ToBeReleased = true
}

func (c *Context) ReleaseTemporaryRegister(reg uint8) {
	c.TemporaryRegistersInUse[reg].InUse = false
	c.TemporaryRegistersInUse[reg].ToBeReleased = true
}

func (c *Context) AddInstruction(instruction *Instruction, freeTmp bool) {
	// if freeTmp is true, then free all registers used by the instruction
	if freeTmp {
		for _, reg := range instruction.RegistersUsed {
			c.ReleaseTemporaryRegister(reg)
		}
	}
	c.Instructions = append(c.Instructions, instruction)
}

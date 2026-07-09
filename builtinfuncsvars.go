package main

import (
	"errors"
	"fmt"
	"gl/yks"
	"log"
	"runtime"
	"strconv"
	"syscall"
	"time"
	"unsafe"

	"github.com/elliotchance/orderedmap/v3"
)

func makeStructObjectFromStructure(structure *yks.Structure, fields []*yks.Field) *yks.StructObject {
	structObject := &yks.StructObject{
		Identifier: structure.Identifier,

		Fields:  fields,
		Methods: []*yks.Method{},

		LastMem: []byte{},
	}

	for _, field := range structure.Fields {
		if field.Method {
			structObject.Methods = append(structObject.Methods, &yks.Method{
				Identifier: "test",
				Func: &yks.Cell{
					FuncValue: field.Func,
				},
			})
		}
	}

	return structObject
}

var (
	gameYKSStructure = &yks.Structure{
		Identifier: "Game",
		Fields: []*yks.FieldDecl{
			{
				Identifier: "test",
				DataType:   "func",
				Method:     true,
				Func: yks.NewFTemp("test", func(v ...any) []any {
					fmt.Println("Testing built in structure")

					return []any{}
				}),
			},
		},
	}

	builtinVals = map[string]any{
		"OS_NAME": func(v ...any) []any {
			return []any{runtime.GOOS}
		},

		"print": func(v ...any) []any {
			fmt.Println(yks.Format(v[yks.BUILTIN_SPECIALS:]...))
			return nil
		},

		"delete": func(v ...any) []any {
			yks.ArgsCheck(v, 2, 2, "table", "any")

			v = v[yks.BUILTIN_SPECIALS:]

			table := v[0].(*orderedmap.OrderedMap[any, any])
			key := v[1]

			table.Delete(key)
			return nil
		},

		"sleep": func(v ...any) []any {
			x, y := v[0].(int), v[1].(int)
			inter := v[2].(*yks.Interpreter)

			if len(v) == 0 {
				throw(inter.CurrentFileName, "Function must have one argument.", x, y)
			}

			v = v[yks.BUILTIN_SPECIALS:]

			switch t := v[0].(type) {
			case float64, float32:
				time.Sleep(time.Duration(yks.MustNTOF64(t) * float64(time.Second)))
			case int64, int32, int16, int8:
				time.Sleep(time.Duration(yks.ToInt64(t) * int64(time.Millisecond)))
			case uint64, uint32, uint16, uint8:
				time.Sleep(time.Duration(yks.ToUint64(t) * uint64(time.Millisecond)))
			default:
				throw(inter.CurrentFileName, "Time value must be a number.", x, y)
			}
			return nil
		},

		"throw": func(v ...any) []any {
			x, y := v[0].(int), v[1].(int)
			inter := v[2].(*yks.Interpreter)

			v = v[yks.BUILTIN_SPECIALS:]
			if len(v) <= 0 {
				throw(inter.CurrentFileName, "Function requires one or more arguments.", x, y)
			}

			throw(inter.CurrentFileName, yks.Format(v...), x, y)
			return nil
		},

		"len": func(v ...any) []any {
			yks.ArgsCheck(v, 1, 1, "any")
			x, y := v[0].(int), v[1].(int)
			inter := v[2].(*yks.Interpreter)

			v = v[yks.BUILTIN_SPECIALS:]

			a := v[0]
			switch a := a.(type) {
			case *yks.Map:
				return []any{int64(a.Len())}
			case string:
				return []any{int64(len(a))}
			case *yks.StructObject:
				layout := a.Layout()
				if len(layout) == 0 {
					return []any{int64(0)}
				}

				lastFieldLayout := layout[len(layout)-1]

				return []any{int64(lastFieldLayout.Offset + lastFieldLayout.Size)}
			default:
				throw(inter.CurrentFileName, "Cannot get lenght of non-string, non-table or non-instance value.", x, y)
			}
			return nil
		},

		"sizeof": func(v ...any) []any {
			yks.ArgsCheck(v, 1, 1, "any")

			v = v[yks.BUILTIN_SPECIALS:]

			a := v[0]
			switch v := a.(type) {
			case *yks.Map:
				a = v.Mem
			}

			return []any{unsafe.Sizeof(a)}
		},

		"time": func(v ...any) []any {
			return []any{time.Now().UnixMilli()}
		},
		"strformat": func(v ...any) []any {
			return []any{yks.Format(v[yks.BUILTIN_SPECIALS:]...)}
		},
		"gettype": func(v ...any) []any {
			yks.ArgsCheck(v, 1, 1, "any")

			v = v[yks.BUILTIN_SPECIALS:]

			return []any{yks.GetValueType(v[0])}
		},
		"numformat": func(v ...any) []any {
			yks.ArgsCheck(v, 2, 2, "string", "bool")
			x, y := v[0].(int), v[1].(int)
			inter := v[2].(*yks.Interpreter)

			v = v[yks.BUILTIN_SPECIALS:]

			str := v[0].(string)
			isint := v[1].(bool)

			if !isint {
				n, err := strconv.ParseFloat(str, 64)
				switch err {
				case strconv.ErrSyntax:
					throw(inter.CurrentFileName, "Syntax error while trying to parse number value.", x, y)
				case strconv.ErrRange:
					throw(inter.CurrentFileName, "Number value is out of range.", x, y)
				}
				return []any{n}
			} else {
				n, err := strconv.ParseInt(str, 0, 64)
				switch err {
				case strconv.ErrSyntax:
					throw(inter.CurrentFileName, "Syntax error while trying to parse number value.", x, y)
				case strconv.ErrRange:
					throw(inter.CurrentFileName, "Number value is out of range.", x, y)
				}

				return []any{n}
			}
		},

		"string": func(v ...any) []any {
			yks.ArgsCheck(v, 1, 1, "table")

			v = v[yks.BUILTIN_SPECIALS:]

			b := v[0].(*yks.Map)
			bstring := []byte{}

		APPEND:
			for _, v := range b.AllFromFront() {
				switch v := v.Get().(type) {
				case int64, int32, int16, int8, uint8, uint16, uint32, uint64:
					charByte := yks.ToUint(yks.ToUint64(v), 8).(byte)

					bstring = append(bstring, charByte)
				default:
					log.Println("Unknown datatype lol the developer is such a shitcoder")
					break APPEND
				}
			}

			return []any{string(bstring)}
		},

		"unicodetostr": func(v ...any) []any {
			yks.ArgsCheck(v, 1, 1, "uint")

			v = v[yks.BUILTIN_SPECIALS:]

			r := rune(yks.ToUint64(v[0]))

			return []any{string(r)}
		},

		"make": func(v ...any) []any {
			yks.ArgsCheck(v, 3, 3, "int", "string", "any")

			x, y := v[0].(int), v[1].(int)
			inter := v[2].(*yks.Interpreter)

			v = v[yks.BUILTIN_SPECIALS:]

			length := int(yks.ToInt64(v[0]))
			dataType := v[1].(string)
			defaultValue := v[2]

			if length < 0 {
				length = 0
			}

			m := &yks.Map{
				OrderedMap: orderedmap.NewOrderedMap[any, *yks.Cell](),

				DataType: dataType,

				Pointers: []any{},
				Layout:   []string{},
				Mem:      []byte{},
			}

			for i := 0; i < length; i++ {
				m.Set(int64(i), yks.CLPTR(inter.CurrentScope, dataType, defaultValue, x, y))
			}
			m.ToMemory()

			return []any{m}
		},

		"cstring": func(v ...any) []any {
			yks.ArgsCheck(v, 1, 1, "string")

			v = v[yks.BUILTIN_SPECIALS:]

			str := v[0].(string)

			slicePtr, err := syscall.BytePtrFromString(str)
			if err == nil {
				err = errors.New("Successfull")
			}

			return []any{
				uintptr(unsafe.Pointer(slicePtr)), err,
			}
		},

		"bytes": func(v ...any) []any {
			yks.ArgsCheck(v, 1, 1, "string")

			x, y := v[0].(int), v[1].(int)
			inter := v[2].(*yks.Interpreter)

			v = v[yks.BUILTIN_SPECIALS:]

			str := v[0].(string)

			slice, err := syscall.ByteSliceFromString(str)
			handle(err)

			m := &yks.Map{
				OrderedMap: orderedmap.NewOrderedMap[any, *yks.Cell](),
				DataType:   "u8",
				Pointers:   []any{},
				Layout:     []string{},
				Mem:        []byte{},
			}

			for i, v := range slice {
				m.Set(int64(i), yks.CLPTR(inter.CurrentScope, "u8", uint8(v), x, y))
			}
			m.ToMemory()

			return []any{
				m,
			}
		},

		"Game": gameYKSStructure,

		"game": makeStructObjectFromStructure(gameYKSStructure, []*yks.Field{}),
	}
)

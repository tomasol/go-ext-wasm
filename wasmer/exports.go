package wasmer

// #include <wasmer_wasm.h>
import "C"
import (
	"fmt"
	"runtime"
	"unsafe"
)

type Exports struct {
	_inner  C.wasm_extern_vec_t
	exports map[string]*Extern
}

func newExports(instance *C.wasm_instance_t, module *Module) *Exports {
	self := &Exports{}
	C.wasm_instance_exports(instance, &self._inner)

	runtime.SetFinalizer(self, func(exports *Exports) {
		C.wasm_extern_vec_delete(exports.inner())
	})

	numberOfExports := int(self.inner().size)
	exports := make(map[string]*Extern, numberOfExports)
	firstExport := unsafe.Pointer(self.inner().data)
	sizeOfExportPointer := unsafe.Sizeof(firstExport)

	var currentExportPointer *C.wasm_extern_t

	moduleExports := module.Exports()

	for nth := 0; nth < numberOfExports; nth++ {
		currentExportPointer = *(**C.wasm_extern_t)(unsafe.Pointer(uintptr(firstExport) + uintptr(nth)*sizeOfExportPointer))
		export := newExtern(currentExportPointer, self)
		exports[moduleExports[nth].Name()] = export
	}

	self.exports = exports

	return self
}

func (self *Exports) inner() *C.wasm_extern_vec_t {
	return &self._inner
}

// Get retrieves and returns an Extern by its name.
//
// ❗️ If the name does not refer to an existing export, Get will return an Error.
//
//   instance, _ := NewInstance(module, NewImportObject())
//   extern, error := instance.Exports.Get("an_export")
//
func (self *Exports) Get(name string) (*Extern, error) {
	export, exists := self.exports[name]

	if exists == false {
		return nil, newErrorWith(fmt.Sprintf("Export `%s` does not exist", name))
	}

	return export, nil
}

// GetRawFunction retrieves and returns a exported Function by its name.
//
// ❗️ If the name does not refer to an existing export, GetRawFunction will return an Error.
//
// ⚠️ If the export is not a function, GetRawFunction will return nil as its result.
//
//   instance, _ := NewInstance(module, NewImportObject())
//   exportedFunc, error := instance.Exports.GetRawFunction("an_exported_function")
//
//   if error != nil && exportedFunc != nil {
//       exportedFunc.Call()
//   }
//
func (self *Exports) GetRawFunction(name string) (*Function, error) {
	exports, err := self.Get(name)

	if err != nil {
		return nil, err
	}

	return exports.IntoFunction(), nil
}

// GetFunction retrieves a exported function by its name and returns it as a native Go function.
//
// ❗️ If the name does not refer to an existing export, GetFunction will return an Error.
//
// ⚠️ If the export is not a function, GetFunction will return nil as its result.
//
//   instance, _ := NewInstance(module, NewImportObject())
//   exportedFunc, error := instance.Exports.GetFunction("an_exported_function")
//
//   if error != nil && exportedFunc != nil {
//       exportedFunc()
//   }
//
func (self *Exports) GetFunction(name string) (func(...interface{}) (interface{}, error), error) {
	function, err := self.GetRawFunction(name)

	if err != nil {
		return nil, err
	}

	return function.Native(), nil
}

// GetGlobal retrieves and returns a exported Global by its name.
//
// ❗️ If the name does not refer to an existing export, GetGlobal will return an Error.
//
// ⚠️ If the export is not a global, GetGlobal will return nil as a result.
//
//   instance, _ := NewInstance(module, NewImportObject())
//   exportedGlobal, error := instance.Exports.GetGlobal("an_exported_global")
//
func (self *Exports) GetGlobal(name string) (*Global, error) {
	exports, err := self.Get(name)

	if err != nil {
		return nil, err
	}

	return exports.IntoGlobal(), nil
}

// GetTable retrieves and returns a exported Table by its name.
//
// ❗️ If the name does not refer to an existing export, GetTable will return an Error.
//
// ⚠️ If the export is not a table, GetTable will return nil as a result.
//
//   instance, _ := NewInstance(module, NewImportObject())
//   exportedTable, error := instance.Exports.GetTable("an_exported_table")
//
func (self *Exports) GetTable(name string) (*Table, error) {
	exports, err := self.Get(name)

	if err != nil {
		return nil, err
	}

	return exports.IntoTable(), nil
}

// GetMemory retrieves and returns a exported Memory by its name.
//
// ❗️ If the name does not refer to an existing export, GetMemory will return an Error.
//
// ⚠️ If the export is not a memory, GetMemory will return nil as a result.
//
//   instance, _ := NewInstance(module, NewImportObject())
//   exportedMemory, error := instance.Exports.GetMemory("an_exported_memory")
//
func (self *Exports) GetMemory(name string) (*Memory, error) {
	exports, err := self.Get(name)

	if err != nil {
		return nil, err
	}

	return exports.IntoMemory(), nil
}

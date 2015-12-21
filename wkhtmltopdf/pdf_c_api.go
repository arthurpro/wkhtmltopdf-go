package wkhtmltopdf

//#cgo CFLAGS: -I/usr/local/include
//#cgo LDFLAGS: -L/usr/local/lib -lwkhtmltox -Wall -ansi -pedantic -ggdb
//#include <stdbool.h>
//#include <stdio.h>
//#include <string.h>
//#include <stdlib.h>
//#include <wkhtmltox/pdf.h>
//extern void progress_changed_cb(void*, const int);
//extern void error_cb(void*, char *msg);
//extern void warning_cb(void*, char *msg);
//extern void phase_changed_cb(void*);
//static void setup_callbacks(wkhtmltopdf_converter * c) {
//  wkhtmltopdf_set_progress_changed_callback(c, (wkhtmltopdf_int_callback)progress_changed_cb);
//  wkhtmltopdf_set_error_callback(c, (wkhtmltopdf_str_callback)error_cb);
//  wkhtmltopdf_set_warning_callback(c, (wkhtmltopdf_str_callback)warning_cb);
//  wkhtmltopdf_set_phase_changed_callback(c, (wkhtmltopdf_void_callback)phase_changed_cb);
//}
import "C"

import (
	"fmt"
	log "github.com/cihub/seelog"
	"unsafe"
)

type GlobalSettings struct {
	s *C.wkhtmltopdf_global_settings
}

type ObjectSettings struct {
	s *C.wkhtmltopdf_object_settings
}

type Converter struct {
	c               *C.wkhtmltopdf_converter
	ProgressChanged func(*Converter, int)
	Error           func(*Converter, string)
	Warning         func(*Converter, string)
	Phase           func(*Converter)
}

var converter_map map[unsafe.Pointer]*Converter

func init() {
	converter_map = map[unsafe.Pointer]*Converter{}
	C.wkhtmltopdf_init(C.false)
}

func NewGolbalSettings() *GlobalSettings {
	return &GlobalSettings{s: C.wkhtmltopdf_create_global_settings()}
}

func (self *GlobalSettings) Set(name, value string) {
	c_name := C.CString(name)
	c_value := C.CString(value)
	defer C.free(unsafe.Pointer(c_name))
	defer C.free(unsafe.Pointer(c_value))
	C.wkhtmltopdf_set_global_setting(self.s, c_name, c_value)
}

func NewObjectSettings() *ObjectSettings {
	return &ObjectSettings{s: C.wkhtmltopdf_create_object_settings()}
}

func (self *ObjectSettings) Set(name, value string) {
	c_name := C.CString(name)
	c_value := C.CString(value)
	defer C.free(unsafe.Pointer(c_name))
	defer C.free(unsafe.Pointer(c_value))
	C.wkhtmltopdf_set_object_setting(self.s, c_name, c_value)
}

func (self *GlobalSettings) NewConverter() *Converter {
	c := &Converter{c: C.wkhtmltopdf_create_converter(self.s)}
	C.setup_callbacks(c.c)

	return c
}

//export progress_changed_cb
func progress_changed_cb(c unsafe.Pointer, p C.int) {
	conv := converter_map[c]
	if conv.ProgressChanged != nil {
		conv.ProgressChanged(conv, int(p))
	}
}

//export error_cb
func error_cb(c unsafe.Pointer, msg *C.char) {
	conv := converter_map[c]
	if conv.Error != nil {
		conv.Error(conv, C.GoString(msg))
	}
}

//export warning_cb
func warning_cb(c unsafe.Pointer, msg *C.char) {
	conv := converter_map[c]
	if conv.Warning != nil {
		conv.Warning(conv, C.GoString(msg))
	}
}

//export phase_changed_cb
func phase_changed_cb(c unsafe.Pointer) {
	conv := converter_map[c]
	if conv.Phase != nil {
		conv.Phase(conv)
	}
}

func (self *Converter) Convert() error {

	// To route callbacks right, we need to save a reference
	// to the converter object, base on the pointer.
	converter_map[unsafe.Pointer(self.c)] = self
	status := C.wkhtmltopdf_convert(self.c)
	delete(converter_map, unsafe.Pointer(self.c))
	if status != C.int(0) {
		return fmt.Errorf("Convert failed")
	}
	return nil
}

func (self *Converter) Add(settings *ObjectSettings) {
	C.wkhtmltopdf_add_object(self.c, settings.s, nil)
}

func (self *Converter) AddHtml(settings *ObjectSettings, data string) {
	c_data := C.CString(data)
	defer C.free(unsafe.Pointer(c_data))
	C.wkhtmltopdf_add_object(self.c, settings.s, c_data)
}

func (self *Converter) Output() (int64, string) {
	cc := C.CString("")
	ccc := (**C.uchar)(unsafe.Pointer(&cc))
	ll := C.wkhtmltopdf_get_output(self.c, ccc)
	co := C.GoStringN((*C.char)(unsafe.Pointer(*ccc)), C.int(ll))
	log.Infof("Converted to %d char.s (%d)", ll, len(co))
	return int64(ll), co
}

func (self *Converter) ErrorCode() int {
	return int(C.wkhtmltopdf_http_error_code(self.c))
}

func (self *Converter) CurrentPhase() (int, string) {
	cpic := C.wkhtmltopdf_current_phase(self.c)
	cpi := int(cpic)
	cps := C.GoString(C.wkhtmltopdf_phase_description(self.c, cpic))
	return cpi, cps
}

func (self *Converter) Destroy() {
	C.wkhtmltopdf_destroy_converter(self.c)
}

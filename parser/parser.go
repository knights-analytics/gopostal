package postal

/*
#cgo pkg-config: libpostal
#include <libpostal/libpostal.h>
#include <stdlib.h>
*/
import "C"

import (
	"log"
	"sync"
	"unicode/utf8"
	"unsafe"
)

var mu sync.Mutex
var libpostalSetup bool

type ParserOptions struct {
	Language string
	Country  string
}

func GetDefaultParserOptions() ParserOptions {
	return ParserOptions{
		Language: "",
		Country:  "",
	}
}

var parserDefaultOptions = GetDefaultParserOptions()

type ParsedComponent struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

func ParseAddressOptions(address string, options ParserOptions) []ParsedComponent {
	if !utf8.ValidString(address) {
		return nil
	}

	mu.Lock()
	defer mu.Unlock()

	if !libpostalSetup {
		if !bool(C.libpostal_setup()) || !bool(C.libpostal_setup_parser()) {
			log.Fatal("Could not load libpostal")
		}
		libpostalSetup = true
	}

	cAddress := C.CString(address)
	defer C.free(unsafe.Pointer(cAddress))

	cOptions := C.libpostal_get_address_parser_default_options()
	if options.Language != "" {
		cLanguage := C.CString(options.Language)
		defer C.free(unsafe.Pointer(cLanguage))

		cOptions.language = cLanguage
	}

	if options.Country != "" {
		cCountry := C.CString(options.Country)
		defer C.free(unsafe.Pointer(cCountry))

		cOptions.country = cCountry
	}

	cAddressParserResponsePtr := C.libpostal_parse_address(cAddress, cOptions)

	if cAddressParserResponsePtr == nil {
		return nil
	}

	cAddressParserResponse := *cAddressParserResponsePtr

	cNumComponents := cAddressParserResponse.num_components
	cComponents := cAddressParserResponse.components
	cLabels := cAddressParserResponse.labels

	numComponents := uint64(cNumComponents)

	parsedComponents := make([]ParsedComponent, numComponents)

	// Accessing a C array
	cComponentsPtr := (*[1 << 30](*C.char))(unsafe.Pointer(cComponents))[:numComponents:numComponents]
	cLabelsPtr := (*[1 << 30](*C.char))(unsafe.Pointer(cLabels))[:numComponents:numComponents]

	var i uint64
	for i = 0; i < numComponents; i++ {
		parsedComponents[i] = ParsedComponent{
			Label: C.GoString(cLabelsPtr[i]),
			Value: C.GoString(cComponentsPtr[i]),
		}
	}

	C.libpostal_address_parser_response_destroy(cAddressParserResponsePtr)

	return parsedComponents
}

func ParseAddress(address string) []ParsedComponent {
	return ParseAddressOptions(address, parserDefaultOptions)
}

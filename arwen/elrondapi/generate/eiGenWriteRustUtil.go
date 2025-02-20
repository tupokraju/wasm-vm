package elrondapigenerate

import (
	"fmt"
	"strings"
)

func rustType(goType string) string {
	if goType == "int32" {
		return "i32"
	}
	if goType == "int64" {
		return "i64"
	}
	return goType
}

func wasmerImportAdapterFunctionName(name string) string {
	return fmt.Sprintf("wasmer_import_%s", snakeCase(name))
}

func cgoFuncPointerFieldName(funcMetadata *EIFunction) string {
	return snakeCase(funcMetadata.Name) + "_func_ptr"
}

func writeRustFnDeclarationArguments(firstArgs string, funcMetadata *EIFunction) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("(%s", firstArgs))
	for _, arg := range funcMetadata.Arguments {
		sb.WriteString(fmt.Sprintf(", %s: %s", snakeCase(arg.Name), rustType(arg.Type)))
	}
	sb.WriteString(")")
	if funcMetadata.Result != nil {
		sb.WriteString(fmt.Sprintf(" -> %s", rustType(funcMetadata.Result.Type)))
	}
	return sb.String()
}

func writeRustFnCallArguments(firstArgs string, funcMetadata *EIFunction) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("(%s", firstArgs))
	for _, arg := range funcMetadata.Arguments {
		sb.WriteString(fmt.Sprintf(", %s", snakeCase(arg.Name)))
	}
	sb.WriteString(")")
	return sb.String()
}

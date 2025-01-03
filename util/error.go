package util

import (
	"errors"
	"fmt"
	"runtime"
)

func ErrorWithInfo(e error, msg string) error {
	_, file, line, ok := runtime.Caller(1) // Caller(1)로 호출한 위치의 정보
	if ok {
		return fmt.Errorf("%s [%s:%d: %w]", msg, file, line, e) // 파일명과 라인 번호 추가
	}
	return e
}

func PrintErrorTrace(e error) {
	for e != nil {
		fmt.Println("Error:", e)
		e = errors.Unwrap(e)
	}
}

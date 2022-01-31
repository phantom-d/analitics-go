package config

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func DynamicCall(obj interface{}, fn string, params ...interface{}) (result interface{}, err error) {
	st := reflect.TypeOf(obj)
	if _, ok := st.MethodByName(fn); !ok {
		return
	}
	method := reflect.ValueOf(obj).MethodByName(fn)
	var inputs []reflect.Value
	if len(params) > 0 {
		for _, v := range params {
			inputs = append(inputs, reflect.ValueOf(v))
		}
	}
	res := method.Call(inputs)
	if res != nil {
		if len(res) > 1 {
			respErr := res[1].Interface()
			if respErr != nil {
				err = respErr.(error)
			}
		}
		result = res[0].Interface()
	}
	return
}

func FmtDuration(d time.Duration) string {
	ms := strings.Split(strconv.FormatFloat(d.Seconds(), 'f', 5, 64), ".")[1]
	d = d.Round(time.Microsecond)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d:%02d:%02d.%s", h, m, s, ms)
}

func GetFuncCurrentName(skip int) (result string) {
	pc := make([]uintptr, 10)
	runtime.Callers(skip, pc)
	f := runtime.FuncForPC(pc[0])
	funcName := f.Name()
	a := strings.Split(funcName, ".")
	result = a[len(a)-1]
	return
}

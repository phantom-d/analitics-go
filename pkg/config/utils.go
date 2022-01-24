package config

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func DynamicCall(obj interface{}, fn string, args map[string]interface{}) (res []reflect.Value) {
	method := reflect.ValueOf(obj).MethodByName(fn)
	var inputs []reflect.Value
	for _, v := range args {
		inputs = append(inputs, reflect.ValueOf(v))
	}
	return method.Call(inputs)
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

func RequestFunc(obj interface{}, name string, skip int, params ...interface{}) (result interface{}, err error) {
	methodName := name + GetFuncCurrentName(skip)
	st := reflect.TypeOf(obj)
	if _, ok := st.MethodByName(methodName); !ok {
		return
	}
	args := make(map[string]interface{}, 0)
	if len(params) > 0 {
		for k, param := range params {
			args["arg"+strconv.Itoa(k)] = param
		}
	}
	res := DynamicCall(obj, methodName, args)
	if len(res) > 1 {
		respErr := res[1].Interface()
		if respErr != nil {
			err = respErr.(error)
		}
	}
	result = res[0].Interface()
	return
}

package main

import (
	"context"
	"fmt"

	//"gitlab.10101111.com/oped/dbms_lib/logrus"
	"github.com/sirupsen/logrus"

	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
)

var (
	rcustomfunc    grpc_recovery.RecoveryHandlerFunc
	rcustomfuncCtx grpc_recovery.RecoveryHandlerFuncContext
)

func RecoveryFromPanic(p interface{}) (err error) {
	logrus.Errorf("panic  recovery from :%v", p)
	return fmt.Errorf("err:%v", p)
}
func RecoveryContexyFromPanic(ctx context.Context, p interface{}) (err error) {
	logrus.Errorf("panic  recovery from :%v", p)
	return fmt.Errorf("ctx:%v err:%v", ctx, p)
}

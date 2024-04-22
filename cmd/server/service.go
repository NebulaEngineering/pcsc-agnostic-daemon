package main

import "context"

type myService struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
}

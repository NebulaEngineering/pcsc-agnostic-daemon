package context

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ebfe/scard"
)

type Context struct {
	*scard.Context
}

// New create a new context to work with readers.
func New() (*Context, error) {

	ctx, err := newContext()
	if err != nil {
		return nil, err
	}

	c := &Context{
		ctx,
	}

	return c, nil
}

func newContext() (*scard.Context, error) {
	ctx, err := scard.EstablishContext()
	if err != nil {
		return nil, err
	}
	if ok, err := ctx.IsValid(); !ok {
		return nil, err
	}
	return ctx, nil
}

// IsValid verify context
func (c *Context) IsValid() (bool, error) {
	if c.Context == nil {
		return false, errors.New("conext is not valid")
	}

	return c.Context.IsValid()

}

// func (c *Context) VerifyContext() error {
// 	if c.Context == nil {
// 		ctx, err := newContext()
// 		if err != nil {
// 			return err
// 		}
// 		c.Context = ctx
// 		return nil
// 	}
// 	if ok, err := c.Context.IsValid(); err != nil || !ok {
// 		fmt.Printf("context is not valid, err: %s\n", err)
// 		ctx, err := newContext()
// 		if err != nil {
// 			return err
// 		}
// 		c.Context = ctx
// 		return nil
// 	}

// 	return nil
// }

// Release release context
func (c *Context) Release() error {

	if c.Context != nil {
		if err := c.Context.Release(); err != nil {
			return err
		}
	}

	return nil
}

// ListReaders list readers detected in context.
func (c *Context) ListReaders() ([]string, error) {

	// if c.Context == nil {
	// 	if ctx, err := scard.EstablishContext(); err != nil {
	// 		return nil, err
	// 	} else {
	// 		c.Context = ctx
	// 	}
	// } else if ok, err := c.Context.IsValid(); err != nil || !ok {
	// 	fmt.Printf("error context: %s, success: %v\n", err, ok)
	// }

	// if c.Context == nil {
	// 	if ctx, err := scard.EstablishContext(); err != nil {
	// 		return nil, err
	// 	} else {
	// 		c.Context = ctx
	// 	}
	// }
	if ok, err := c.IsValid(); err != nil || !ok {
		return nil, fmt.Errorf("context is not valid, err: %w", err)
	}

	readers, err := c.Context.ListReaders()
	if err != nil {
		// c.Context = nil
		return nil, err
	}

	rds := make([]string, 0)
	rds = append(rds, readers...)
	return rds, nil
}

// ReaderInformation verify reader with regex "key" and return real name's reader.
func (c *Context) ReaderInformation(key string) (string, error) {

	if ok, err := c.IsValid(); err != nil || !ok {
		return "", fmt.Errorf("context is not valid, err: %w", err)
	}
	readers, err := c.Context.ListReaders()
	if err != nil {
		// c.Context = nil
		return "", err
	}

	for _, r := range readers {
		if strings.Contains(r, key) {
			return r, nil
		}
	}
	return "", nil
}
